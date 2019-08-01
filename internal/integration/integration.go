package integration

// Handler kinds
const (
	HTTP        = "HTTP"
	InfluxDB    = "INFLUXDB"
	ThingsBoard = "THINGSBOARD"
)

// Integrator defines the interface that an intergration must implement.
type Integrator interface {
	SendDataUp(pl DataUpPayload) error                      // send data-up payload，数据上传
	SendJoinNotification(pl JoinNotification) error         // send join notification，设备加入通知
	SendACKNotification(pl ACKNotification) error           // send ack notification，回应通知
	SendErrorNotification(pl ErrorNotification) error       // send error notification，错误通知
	SendStatusNotification(pl StatusNotification) error     // send status notification，状态通知
	SendLocationNotification(pl LocationNotification) error // send location notification，定位通知
	DataDownChan() chan DataDownPayload                     // returns DataDownPayload channel，数据下发通道
	Close() error                                           // closes the handler，关闭请求
}

var integration Integrator

// Integration returns the integration object.
func Integration() Integrator {
	if integration == nil {
		panic("integration package must be initialized")
	}
	return integration
}

// SetIntegration sets the given integration.
func SetIntegration(i Integrator) {
	integration = i
}
