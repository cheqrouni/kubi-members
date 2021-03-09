package ldap

import (
	"crypto/tls"
	"fmt"
	"github.com/ca-gip/kubi-members/internal/utils"
	cagipv1 "github.com/ca-gip/kubi-members/pkg/apis/ca-gip/v1"
	ldap "github.com/go-ldap/ldap/v3"
	"k8s.io/klog/v2"
	"syscall"
)

type User struct {
	Dn          string `ldap:"dn"`
	Cn          string `ldap:"cn"`
	DisplayName string `ldap:"displayName"`
	Mail        string `ldap:"mail"`
}

func (u *User) toProjectMember() *cagipv1.ProjectMember {
	return &cagipv1.ProjectMember{
		Dn:          u.Dn,
		Cn:          u.Cn,
		DisplayName: u.DisplayName,
		Mail:        u.Mail,
	}
}

type Ldap struct {
	Conn       *ldap.Conn
	UserBase   string
	UserFilter string
	GroupBase  string
}

func NewLdap() *Ldap {

	config := utils.LoadConfig()

	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: config.SkipTLSVerification,
	}

	var (
		err  error
		conn *ldap.Conn
	)

	if config.UseSSL {
		conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	}

	if config.StartTLS {
		err = conn.StartTLS(tlsConfig)
		if err != nil {
			klog.Fatal("unable to setup TLS connection")
			syscall.Exit(1)
		}
	}

	if err != nil {
		klog.Fatalf("unable to create ldap connector for %s:%d", config.Host, config.Port)
		syscall.Exit(1)
	}

	// Bind with BindAccount
	err = conn.Bind(config.BindDN, config.BindPassword)

	if err != nil {
		klog.Fatalf("Error while binding : %s", err)
		syscall.Exit(1)
	}

	return &Ldap{Conn: conn, UserBase: config.UserBase, GroupBase: config.GroupBase}

}

func (l *Ldap) searchUser(userDN string) (user *User, err error) {
	user = &User{}
	res, err := l.Conn.Search(&ldap.SearchRequest{
		BaseDN:       userDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    1,
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       "(|(objectClass=person)(objectClass=organizationalPerson))",
		Attributes:   []string{"cn", "mail", "displayName"},
	})

	if err != nil || res == nil || len(res.Entries) == 0 {
		return
	} else {
		err = utils.Unmarshal(res.Entries[0], user)
		return
	}
}
