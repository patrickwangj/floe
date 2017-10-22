package config

import (
	"fmt"

	"gopkg.in/yaml.v2"

	nt "github.com/floeit/floe/config/nodetype"
	"github.com/floeit/old/log"
)

// Config is the set of nodes and rules
type Config struct {
	Common commonConfig
	// the list of flow configurations
	Flows []*Flow
}

type commonConfig struct {
	// all other floe Hosts
	Hosts []string
	// the api base url
	BaseURL string `yaml:"base-url"`
}

// FoundFlow is a struct containing a cut down set of properties of a Flow.
// It can be used to decide on the best host to use to run this Flow.
type FoundFlow struct {
	Ref          FlowRef
	ReuseSpace   bool
	ResourceTags []string
	HostTags     []string

	Nodes []Node
}

// FindFlowsBySubs finds all flows where its subs match the given params
func (c *Config) FindFlowsBySubs(eType string, flow *FlowRef, opts nt.Opts) map[FlowRef]FoundFlow {
	res := map[FlowRef]FoundFlow{}
	for _, f := range c.Flows {
		// if a flow is specified it has to match
		if flow != nil {
			log.Debugf("config - comparing flow:<%s> to config flow:<%s-%s>", flow, f.ID, f.Ver)
			if f.ID != flow.ID || f.Ver != flow.Ver {
				continue
			}
		}
		log.Debugf("config - found flow: <%s-%s>. %d triggers", f.ID, f.Ver, len(f.Triggers))
		// match on other stuff
		ns := f.matchTriggers(eType, &opts)
		// found some matching nodes for this flow
		if len(ns) > 0 {
			// make sure this flow is in the results
			fr := ns[0].FlowRef()
			ff, ok := res[fr]
			if !ok {
				ff = FoundFlow{
					Ref:          fr,
					ReuseSpace:   f.ReuseSpace,
					ResourceTags: f.ResourceTags,
					HostTags:     f.HostTags,
				}
			}
			ff.Nodes = []Node{}
			for _, n := range ns {
				ff.Nodes = append(ff.Nodes, n)
			}
			res[fr] = ff
		} else {
			log.Debugf("config - flow:<%s> failed on trigger match", flow)
		}
	}
	return res
}

// FindFlow finds the specific flow where its subs match the given params
func (c *Config) FindFlow(f FlowRef, eType string, opts nt.Opts) (FoundFlow, bool) {
	found := c.FindFlowsBySubs(eType, nil, opts)
	flow, ok := found[f]
	return flow, ok
}

// Flow returns the flow config matching the id and version
func (c *Config) Flow(fRef FlowRef) *Flow {
	for _, f := range c.Flows {
		if f.ID == fRef.ID && f.Ver == fRef.Ver {
			return f
		}
	}
	return nil
}

// LatestFlow returns the flow config matching the id with the highest version
func (c *Config) LatestFlow(id string) *Flow {
	var latest *Flow
	highestVer := 0
	for _, f := range c.Flows {
		if f.ID != id {
			continue
		}
		if f.Ver > highestVer {
			latest = f
		}
	}
	return latest
}

// FindNodeInFlow returns the nodes matching the tag in this flow matching the id and version
func (c *Config) FindNodeInFlow(fRef FlowRef, tag string) (FoundFlow, bool) {
	ff := FoundFlow{}
	f := c.Flow(fRef)
	if f == nil {
		return ff, false
	}
	// we found the matching flow so can find any matching nodes
	ff = FoundFlow{
		Ref:          fRef,
		ReuseSpace:   f.ReuseSpace,
		ResourceTags: f.ResourceTags,
		HostTags:     f.HostTags,
		Nodes:        []Node{},
	}
	// normal tasks
	ns := f.matchTag(NcTask, tag)
	for _, n := range ns {
		ff.Nodes = append(ff.Nodes, Node(n))
	}
	// merge nodes
	ns = f.matchTag(NcMerge, tag)
	for _, n := range ns {
		ff.Nodes = append(ff.Nodes, Node(n))
	}
	return ff, true
}

// zero sets up all the default values
func (c *Config) zero() error {
	for i, f := range c.Flows {
		if err := f.zero(); err != nil {
			return fmt.Errorf("flow %d - %v", i, err)
		}
	}
	return nil
}

// ParseYAML takes a YAML input as a byte array and returns a Config object
// or an error
func ParseYAML(in []byte) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal(in, &c)
	if err != nil {
		return c, err
	}
	err = c.zero()
	return c, err
}
