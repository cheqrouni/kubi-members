package ldap

import (
	"crypto/tls"
	"fmt"
	"github.com/ca-gip/kubi-members/internal/utils"
	ldap "github.com/go-ldap/ldap/v3"
	"k8s.io/klog/v2"
	"syscall"
)

type User struct {
	Dn       string `ldap:"dn"`
	Username string `ldap:"displayName"`
	Mail     string `ldap:"mail"`
}

type Users []User

func (u Users) Exist(dn string) bool {
	for _, user := range u {
		if user.Dn == dn {
			return true
		}
	}
	return false
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

func (l *Ldap) searchGroupMember(groupDN string) (members []string, err error) {
	res, err := l.Conn.Search(&ldap.SearchRequest{
		BaseDN:       groupDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    30,
		TypesOnly:    false,
		Filter:       "(|(objectClass=groupOfNames)(objectClass=group))",
		Attributes:   []string{"member"},
	})

	if err != nil || res == nil || len(res.Entries) != 1 {
		return
	}

	members = res.Entries[0].GetAttributeValues("member")

	return
}

func (l *Ldap) searchUser(userDN string, usernameAttr string, mailAttr string) (user *User, err error) {
	res, err := l.Conn.Search(&ldap.SearchRequest{
		BaseDN:       userDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    1,
		TimeLimit:    10,
		TypesOnly:    false,
		Filter:       "(|(objectClass=person)(objectClass=organizationalPerson))",
		Attributes:   []string{usernameAttr, mailAttr},
	})

	if err != nil || res == nil || len(res.Entries) == 0 {
		return
	} else {
		user = &User{
			Dn:       userDN,
			Username: res.Entries[0].GetAttributeValue(usernameAttr),
			Mail:     res.Entries[0].GetAttributeValue(mailAttr),
		}
		return
	}
}

func (l *Ldap) Search(groupDN string) (users Users, err error) {
	membersDn, err := l.searchGroupMember(groupDN)
	if err != nil {
		return
	}

	for _, memberDn := range membersDn {
		user, _ := l.searchUser(memberDn, "cn", "mail")
		if user != nil {
			users = append(users, *user)
		}
	}

	return
}
