package constant

// QueueName Rabbit Config Const
const QueueName = "Order.Projection"
const RoutingKey = "Order.*.*"

// OrderLineCreatedExchange Const
const OrderLineCreatedExchange = "Order.Events.V1.OrderLine:Created"
const OrderLineInProgressedExchange = "Order.Events.V1.OrderLine:InProgressed"
const OrderLineInTransittedExchange = "Order.Events.V1.OrderLine:InTransitted"
const OrderLineDeliveredExchange = "Order.Events.V1.OrderLine:Delivered"

const qaConfigPath = "C:\\Users\\musta\\OneDrive\\Masaüstü\\event-ordering-consumers-rabbitMq-with-consistent-hash\\tools\\k8s-files\\qa\\config.qa.json"
