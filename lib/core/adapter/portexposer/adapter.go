package portexposer

type PortExposer interface {
	Start()
	Stop()
	Port() uint16
}
