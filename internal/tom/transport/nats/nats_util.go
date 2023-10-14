package nats

import "strings"

func generateConsumerMetadata(consumerkey string) (durable, queuegroup, deliversubj string) {
	return validDurableName("durable_" + consumerkey), "queue." + consumerkey, "deliver." + consumerkey
}

func validDurableName(somename string) string {
	return strings.ReplaceAll(somename, ".", "_")
}
