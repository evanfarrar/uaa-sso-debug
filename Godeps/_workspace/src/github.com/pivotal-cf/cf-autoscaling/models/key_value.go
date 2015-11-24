package models

import "database/sql"

type KeyValueInterface interface {
	Set(string, string) error
	Get(string) (string, error)
}

type KeyValue struct{}

type kv struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

func NewKeyValueRepo() KeyValue {
	return KeyValue{}
}

func (repo KeyValue) Get(key string) (string, error) {
	kv := kv{}

	err := Database().Connection.SelectOne(&kv, "SELECT * FROM `key_value_store` WHERE `key` = ?", key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		} else {
			return "", err
		}
	}

	return kv.Value, nil
}

func (repo KeyValue) Set(key, value string) error {
	keyValue := kv{
		Key:   key,
		Value: value,
	}

	value, err := repo.Get(key)
	if err != nil {
		return err
	}

	if value == "" {
		err = Database().Connection.Insert(&keyValue)
	} else {
		_, err = Database().Connection.Update(&keyValue)
	}

	return err
}
