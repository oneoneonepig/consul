package export

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"

	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/flags"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}

type cmd struct {
	UI    cli.Ui
	flags *flag.FlagSet
	http  *flags.HTTPFlags
	help  string

	serviceName    string
	peerNames      string
	partitionNames string
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)

	c.flags.StringVar(&c.serviceName, "name", "", "(Required) Specify the name of the service you want to export.")
	c.flags.StringVar(&c.peerNames, "consumer-peers", "", "(Required) Peers the service will be exported to, formatted as a comma-separated list. Not required for Enterprise if setting -consumer-partitions.")
	c.flags.StringVar(&c.partitionNames, "consumer-partitions", "", "Required if not setting -consumer-peers. The local partitions within the same datacenter that the service will be exported to, formatted as a comma-separated list. Admin Partitions are a Consul Enterprise feature.")

	c.http = &flags.HTTPFlags{}
	flags.Merge(c.flags, c.http.ClientFlags())
	flags.Merge(c.flags, c.http.MultiTenancyFlags())
	c.help = flags.Usage(help, c.flags)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if err := c.validateFlags(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	var peerNames []string
	if c.peerNames != "" {
		peerNames = strings.Split(c.peerNames, ",")
		for _, peerName := range peerNames {
			if peerName == "" {
				c.UI.Error(fmt.Sprintf("Invalid peer %q", peerName))
				return 1
			}
		}
	}

	partitionNames, err := c.getPartitionNames()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	client, err := c.http.APIClient()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error connect to Consul agent: %s", err))
		return 1
	}

	entry, _, err := client.ConfigEntries().Get("exported-services", "default", nil)
	if err != nil && !strings.Contains(err.Error(), agent.ConfigEntryNotFoundErr) {
		c.UI.Error(fmt.Sprintf("Error reading config entry %s/%s: %v", "exported-services", "default", err))
		return 1
	}

	var cfg *api.ExportedServicesConfigEntry
	if entry == nil {
		cfg = c.initializeConfigEntry(peerNames, partitionNames)
	} else {
		cfg, ok := entry.(*api.ExportedServicesConfigEntry)
		if !ok {
			c.UI.Error(fmt.Sprintf("Existing config entry has incorrect type: %t", entry))
			return 1
		}

		cfg = c.updateConfigEntry(cfg, peerNames, partitionNames)
	}

	ok, _, err := client.ConfigEntries().CAS(cfg, cfg.GetModifyIndex(), nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error writing config entry: %s", err))
		return 1
	} else if !ok {
		c.UI.Error(fmt.Sprintf("Config entry was changed during update. Please try again"))
		return 1
	}

	switch {
	case len(c.peerNames) > 0 && len(c.partitionNames) > 0:
		c.UI.Info(fmt.Sprintf("Successfully exported service %q to peers %q and to partitions %q", c.serviceName, c.peerNames, c.partitionNames))
	case len(c.peerNames) > 0:
		c.UI.Info(fmt.Sprintf("Successfully exported service %q to peers %q", c.serviceName, c.peerNames))
	case len(c.partitionNames) > 0:
		c.UI.Info(fmt.Sprintf("Successfully exported service %q to partitions %q", c.serviceName, c.partitionNames))
	}

	return 0
}

func (c *cmd) initializeConfigEntry(peerNames, partitionNames []string) *api.ExportedServicesConfigEntry {
	return &api.ExportedServicesConfigEntry{
		Name: "default",
		Services: []api.ExportedService{
			{
				Name:      c.serviceName,
				Consumers: buildConsumers(peerNames, partitionNames),
			},
		},
	}
}

func (c *cmd) updateConfigEntry(cfg *api.ExportedServicesConfigEntry, peerNames, partitionNames []string) *api.ExportedServicesConfigEntry {
	serviceExists := false

	for i, service := range cfg.Services {
		if service.Name == c.serviceName {
			serviceExists = true

			// Add a consumer for each peer where one doesn't already exist
			for _, peerName := range peerNames {
				peerExists := false
				for _, consumer := range service.Consumers {
					if consumer.Peer == peerName {
						peerExists = true
						break
					}
				}
				if !peerExists {
					cfg.Services[i].Consumers = append(cfg.Services[i].Consumers, api.ServiceConsumer{Peer: peerName})
				}
			}

			// Add a consumer for each partition where one doesn't already exist
			for _, partitionName := range partitionNames {
				partitionExists := false

				for _, consumer := range service.Consumers {
					if consumer.Partition == partitionName {
						partitionExists = true
						break
					}
				}
				if !partitionExists {
					cfg.Services[i].Consumers = append(cfg.Services[i].Consumers, api.ServiceConsumer{Partition: partitionName})
				}
			}
		}
	}

	if !serviceExists {
		cfg.Services = append(cfg.Services, api.ExportedService{
			Name:      c.serviceName,
			Consumers: buildConsumers(peerNames, partitionNames),
		})
	}

	return cfg
}

func buildConsumers(peerNames []string, partitionNames []string) []api.ServiceConsumer {
	var consumers []api.ServiceConsumer
	for _, peer := range peerNames {
		consumers = append(consumers, api.ServiceConsumer{
			Peer: peer,
		})
	}
	for _, partition := range partitionNames {
		consumers = append(consumers, api.ServiceConsumer{
			Partition: partition,
		})
	}
	return consumers
}

//========

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return flags.Usage(c.help, nil)
}
