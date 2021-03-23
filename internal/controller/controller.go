package controller

import (
	"github.com/ca-gip/kubi-members/internal/ldap"
	"github.com/ca-gip/kubi-members/internal/utils"
	v1 "github.com/ca-gip/kubi-members/pkg/apis/ca-gip/v1"
	membersclientset "github.com/ca-gip/kubi-members/pkg/generated/clientset/versioned"
	kubiv1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	projectclientset "github.com/ca-gip/kubi/pkg/generated/clientset/versioned"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type Controller struct {
	configmapclientset kubernetes.Interface
	projectclientset   projectclientset.Interface
	membersclientset   membersclientset.Interface
	projectsMembers		map[string][]*v1.ProjectMember
	clusterMembers     []*v1.ClusterMember

	ldap *ldap.Ldap
}

func NewController(configMapClient kubernetes.Interface, projectClient projectclientset.Interface, membersClient membersclientset.Interface, ldap *ldap.Ldap) *Controller {
	return &Controller{
		configmapclientset: configMapClient,
		projectclientset:   projectClient,
		membersclientset:   membersClient,
		ldap:               ldap,
	}
}

func (c *Controller) Preflight() {
}

func (c *Controller) Run() (err error) {

	c.clusterMembers = []*v1.ClusterMember{}
	c.projectsMembers = make(map[string][]*v1.ProjectMember)

	err = c.LocalSyncClusterMembers()
	if err != nil {
		klog.Fatalf("Could not local compute cluster members : %v", err)
		return
	}

	err = c.LocalSyncProjectsMembers()
	if err != nil {
		klog.Fatalf("Could not local compute project members: %v", err)
		return
	}

	c.SyncClusterMembers()
	c.SyncProjectMembers()

	klog.Infof("Update members job complete.")

	return
}

func (c *Controller) SyncClusterMembers() {
	c.membersclientset.CagipV1().ClusterMembers().DeleteCollection(&metav1.DeleteOptions{},metav1.ListOptions{})

	for _, member := range c.clusterMembers {
		_, err := c.membersclientset.CagipV1().ClusterMembers().Create(member)
		if err != nil {
			klog.Errorf("Could not create cluster member %s", member.Username, err)
		}
	}
}

func (c *Controller) SyncProjectMembers() {
	c.clearProjectsMembers()
	for project, members := range c.projectsMembers {
		c.createProjectMembers(project,members)
	}

}


func (c *Controller) LocalSyncProjectsMembers() error {
	projects, err := c.projectclientset.CagipV1().Projects().List(metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Could not list project : %s", err)
		return err
	}
	for _, project := range projects.Items {
		if project.Status.Name == kubiv1.ProjectStatusCreated {
			members, err := c.ldap.Search(project.Spec.SourceDN)
			if err != nil {
				klog.Errorf("Could not find ldap members for %s : %s", project.Spec.SourceDN, err)
			}
			c.projectsMembers[project.Name] = c.templateProjectMembers(&project, members)
		}
	}
	return nil
}


func (c *Controller) LocalSyncClusterMembers() error {

	opsUsers, err := c.ldap.Search(c.ldap.OpsGroupBase)
	if err != nil {
		klog.Errorf("Could not find ldap members for %s : %s", c.ldap.OpsGroupBase, err)
	}
	c.synchronizeClusterMembersByRole(opsUsers, utils.OpsRole)

	appUsers, err := c.ldap.Search(c.ldap.AppGroupBase)
	if err != nil {
		klog.Errorf("Could not find ldap members for %s : %s", c.ldap.AppGroupBase, err)
	}
	c.synchronizeClusterMembersByRole(appUsers, utils.AppRole)

	customerUsers, err := c.ldap.Search(c.ldap.CustomerGroupBase)
	if err != nil {
		klog.Errorf("Could not find ldap members for %s : %s", c.ldap.CustomerGroupBase, err)
	}
	c.synchronizeClusterMembersByRole(customerUsers, utils.CustomerRole)

	adminsUsers, err := c.ldap.Search(c.ldap.AdminGroupBase)
	if err != nil {
		klog.Errorf("Could not find ldap members for %s : %s", c.ldap.CustomerGroupBase, err)
	}
	c.synchronizeClusterMembersByRole(adminsUsers, utils.AdminRole)

	return nil
}



func (c *Controller) indexOfClusterMember(user ldap.User) int {
	for i := 0; i < len(c.clusterMembers); i++ {
		if c.clusterMembers[i].Mail == user.Mail {
			return i
		}
	}
	return -1
}

func (c *Controller) synchronizeClusterMembersByRole(members ldap.Users, role utils.ClusterRole) {
	for _, member := range members {
		userIndex := c.indexOfClusterMember(member)
		if userIndex == -1 {
			c.clusterMembers = append(c.clusterMembers, c.templateClusterMember(member, role))
		} else {
			// Change user Role only if current has less privileges
			_, userRole := utils.GetClusterRole(c.clusterMembers[userIndex].Role)
			if userRole < role {
				c.clusterMembers[userIndex].Role = role.String()
			}
		}
	}
}

func (c *Controller) templateClusterMember(member ldap.User, role utils.ClusterRole) *v1.ClusterMember {
	return &v1.ClusterMember{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ToDNSString(member.Mail),
		},
		Dn:       member.Dn,
		Username: member.Username,
		Mail:     member.Mail,
		Role:     role.String(),
	}
}

func (c *Controller) createProjectMembers(namespace string, members []*v1.ProjectMember) {
	for _, member := range members {
		_, err := c.membersclientset.CagipV1().ProjectMembers(namespace).Get(member.Username, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			_, err := c.membersclientset.CagipV1().ProjectMembers(namespace).Create(member)
			if err != nil {
				klog.Errorf("Could not create ProjectMember %s : %s", member.Username, err)
			}
		}
	}
	return
}

func (c *Controller) templateProjectMember(project *kubiv1.Project, user ldap.User) *v1.ProjectMember {
	return &v1.ProjectMember{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.ToDNSString(user.Mail),
			Namespace: project.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(project, kubiv1.SchemeGroupVersion.WithKind("Project")),
			},
		},
		Dn:       user.Dn,
		Username: user.Username,
		Mail:     user.Mail,
	}
}

func (c *Controller) templateProjectMembers(project *kubiv1.Project, users []ldap.User) (members []*v1.ProjectMember) {
	for _, user := range users {
		member := c.templateProjectMember(project, user)
		members = append(members, member)
	}
	return
}

func (c *Controller) clearProjectsMembers() {
	projects, err := c.projectclientset.CagipV1().Projects().List(metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Could not list projects")
	}
	for _, project := range projects.Items {
		err = c.membersclientset.CagipV1().ProjectMembers(project.Name).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Could not remove members from project %s: %v", project.Name, err)
		}
	}
}
