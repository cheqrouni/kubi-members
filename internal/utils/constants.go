package utils

const (
	ControllerName = "kotary-controller"

	CouldNotList = "Could not list resources %s"


)

type ClusterRole int

const (
	CustomerRole 	  ClusterRole	= iota
	AppRole
	OpsRole
	AdminRole
)

func (c ClusterRole) String() string{
	return  [...]string{"CustomerOps", "AppOps", "ClusterOps", "Admin"}[c]
}


