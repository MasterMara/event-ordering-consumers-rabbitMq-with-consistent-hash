package constant

// QueueName Rabbit Config Const
const QueueName = "Order.Projection"
const RoutingKey = "Order.Master.İstanbul"

// OrderLineCreatedExchange Const
const OrderLineCreatedExchange = "Order.Events.V1.Order:Created"
const OrderLineInProgressedExchange = "Order.Events.V1.Order:InProgressed"
const OrderLineInTransittedExchange = "Order.Events.V1.Order:InTransitted"
const OrderLineDeliveredExchange = "Order.Events.V1.Order:Delivered"

const qaConfigPath = "C:\\Users\\musta\\OneDrive\\Masaüstü\\event-ordering-consumers-rabbitMq-with-consistent-hash\\tools\\k8s-files\\qa\\config.qa.json"
