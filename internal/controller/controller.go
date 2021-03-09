package controller

import (
	"fmt"
	"github.com/ca-gip/kubi-members/internal/ldap"
	cagipv1 "github.com/ca-gip/kubi-members/pkg/apis/ca-gip/v1"
	membersclientset "github.com/ca-gip/kubi-members/pkg/generated/clientset/versioned"
	kubiv1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	projectclientset "github.com/ca-gip/kubi/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Controller struct {
	configmapclientset kubernetes.Interface
	projectclientset   projectclientset.Interface
	membersclientset   membersclientset.Interface
	ldap               *ldap.Ldap
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
	projects, err := c.projectclientset.CagipV1().Projects().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, project := range projects.Items {
		c.SyncMembers(project)
	}
	return
}

func (c *Controller) SyncMembers(project kubiv1.Project) {
	fmt.Println(project)

}

func (c *Controller) newProjectMembers(namespace string, members cagipv1.ProjectMembers) *cagipv1.ProjectMembers {
	return &cagipv1.ProjectMembers{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespace,
			Namespace: namespace,
		},
		// Members: members
	}
}
