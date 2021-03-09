module github.com/ca-gip/kubi-members

go 1.16

require (
	github.com/ca-gip/kubi v1.8.6
	github.com/go-ldap/ldap/v3 v3.2.4
	github.com/joho/godotenv v1.3.0
	k8s.io/api v0.20.4 // indirect
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/code-generator v0.20.4
	k8s.io/klog/v2 v2.4.0
)

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
