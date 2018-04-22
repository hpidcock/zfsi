package config_stream

import (
	"bytes"
	"log"

	"github.com/cnf/structhash"
	"github.com/hpidcock/zfsi/pkg/service-agg"
)

type Config []service_agg.Service

type ConfigSubscription struct {
	ConfigChan chan Config
	closeChan  chan interface{}
}

func (sub *ConfigSubscription) Close() {
	close(sub.closeChan)
	close(sub.ConfigChan)
}

type ConfigStream struct {
	closeChan  chan interface{}
	closedChan chan interface{}

	newSubsciptions chan *ConfigSubscription
	configChan      chan Config

	subscriptions []*ConfigSubscription

	lastConfig     Config
	lastConfigHash []byte
}

func NewConfigStream() *ConfigStream {
	cs := &ConfigStream{}
	cs.closeChan = make(chan interface{})
	cs.closedChan = make(chan interface{})
	cs.configChan = make(chan Config)
	cs.newSubsciptions = make(chan *ConfigSubscription, 10)
	cs.subscriptions = make([]*ConfigSubscription, 0)
	go cs.run()
	return cs
}

func (cs *ConfigStream) Register() *ConfigSubscription {
	sub := &ConfigSubscription{
		ConfigChan: make(chan Config, 1),
		closeChan:  make(chan interface{}),
	}

	cs.newSubsciptions <- sub
	return sub
}

func (cs *ConfigStream) Publish(config Config) {
	cs.configChan <- config
}

func (cs *ConfigStream) Close() {
	close(cs.closeChan)
	<-cs.closedChan

out:
	for {
		select {
		case subsciption := <-cs.newSubsciptions:
			close(subsciption.ConfigChan)
			close(subsciption.closeChan)
		default:
			break out
		}
	}

	close(cs.newSubsciptions)

	for _, subsciption := range cs.subscriptions {
		close(subsciption.ConfigChan)
		close(subsciption.closeChan)
	}
	cs.subscriptions = nil

	close(cs.configChan)
}

func (cs *ConfigStream) run() {
	log.Print("start configstream")
	defer log.Print("end configstream")
	defer close(cs.closedChan)
	for {
		select {
		case <-cs.closeChan:
			return
		case config := <-cs.configChan:
			cs.broadcast(config)
		case subsciption := <-cs.newSubsciptions:
			subsciption.ConfigChan <- cs.lastConfig
			cs.subscriptions = append(cs.subscriptions, subsciption)
		}
	}
}

func (cs *ConfigStream) broadcast(config Config) {
	hash := structhash.Md5(config, 1)
	if bytes.Compare(hash, cs.lastConfigHash) == 0 {
		return
	}

	cs.lastConfig = config
	cs.lastConfigHash = hash

	published := make([]*ConfigSubscription, 0)
	for _, subscription := range cs.subscriptions {
		select {
		case <-subscription.closeChan:
			continue
		default:
			subscription.ConfigChan <- config
			published = append(published, subscription)
		}
	}
	cs.subscriptions = published
}
