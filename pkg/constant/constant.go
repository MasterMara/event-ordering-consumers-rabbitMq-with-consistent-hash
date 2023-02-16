package constant

// QueueName Rabbit Config Const
const QueueName = ""
const RoutingKey = ""

// OrderLineCreatedExchange Const
const OrderLineCreatedExchange = "Order.Events.V1.OrderLine:Created"
const OrderLineInProgressedExchange = "Order.Events.V1.OrderLine:InProgressed"
const OrderLineInTransittedExchange = "Order.Events.V1.OrderLine:InTransitted"
const OrderLineDeliveredExchange = "Order.Events.V1.OrderLine:Delivered"
