package etcd_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/codec/yaml"
	"github.com/DoNewsCode/core/config/remote/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"strings"
	"time"
)

func Example() {
	addr := os.Getenv("ETCD_ADDR")
	if addr == "" {
		fmt.Println("set ETCD_ADDR for run example")
		return
	}
	key := "core.yaml"
	envEtcdAddrs := strings.Split(addr, ",")
	cfg := clientv3.Config{
		Endpoints:   envEtcdAddrs,
		DialTimeout: time.Second,
	}
	_ = put(cfg, key, "name: etcd")

	c := core.New(etcd.WithKey(cfg, key, yaml.Codec{}))
	c.ProvideEssentials()
	fmt.Println(c.String("name"))

	// Output:
	// etcd
}

func put(cfg clientv3.Config, key, val string) error {
	client, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.Put(ctx, key, val)
	if err != nil {
		return err
	}

	return nil
}
