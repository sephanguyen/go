package bootstrap

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	commonFieldName     = "Common"
	postgresV2FieldName = "PostgresV2"
	listenerFieldName   = "Listener"
	natsjsFieldName     = "NatsJS"
	kafkaFieldName      = "KafkaCluster"
	elasticFieldName    = "ElasticSearch"
	unleashFieldName    = "UnleashClientConfig"
	storageFieldName    = "Storage"
)

var errFieldNotFound = errors.New("field not found")

func ignoreErrFieldNotFound(err error) error {
	if errors.Is(err, errFieldNotFound) {
		return nil
	}
	return err
}

func extract[T any](config interface{}, field string) (*T, error) {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct")
	}
	fv := v.FieldByName(field)
	if !fv.IsValid() {
		return nil, errFieldNotFound
	}
	out, ok := fv.Interface().(T)
	if !ok {
		tv := fv.Type()
		return nil, fmt.Errorf("expected field %q to be of \"%T\" type, got \"%s.%s\"", field, out, tv.PkgPath(), tv.Name())
	}
	return &out, nil
}

func Extract[T any](config interface{}, field string) (*T, error) {
	return extract[T](config, field)
}
