package pubsub

const (
	SalesDiscoveryPort = 41000 + iota
	SoldToDiscoveryPort
	BidDiscoveryPort
	AckDiscoveryPort
	OrderDeliveredDiscoveryPort
	DbDiscoveryPort
)

const SalesTopic = "sales"
const SoldToTopic = "sold to"
const BidTopic = "bid"
const AckTopic = "ack"
const DbDiscoveryTopic = "db"
const OrderDeliveredTopic = "order del"
