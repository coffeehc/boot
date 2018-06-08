package loadbalancer

type NodeDown interface {
	Delete(addr Address)
}
