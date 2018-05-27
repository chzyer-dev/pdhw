package pdhw

import (
	"bytes"
	"fmt"
	"strings"
)

type Stores []Store

func (s Stores) Count(key string, value string) int {
	var val int
	for idx := range s {
		if s[idx].GetLabel(key) == value {
			val++
		}
	}
	return val
}

func (s Stores) Get(id int) *Store {
	idx := s.Index(id)
	if idx >= 0 {
		return &s[idx]
	}
	return nil
}

func (s Stores) Index(id int) int {
	for idx := range s {
		if s[idx].ID == id {
			return idx
		}
	}
	return -1
}

type Store struct {
	ID     int
	Labels map[string]string
}

func (s *Store) IsValid() bool {
	return s.ID > 0
}

func (s *Store) SetLabel(key string, value string) {
	if s.Labels == nil {
		s.Labels = make(map[string]string, 8)
	}
	s.Labels[key] = value
}

func (s *Store) GetLabel(key string) string {
	return s.Labels[key]
}

// ---------------------------------------------------------------------------

type StoreTree struct {
	Datacenters []StoreTreeDatacenter
}

func NewStoreTree() *StoreTree {
	return &StoreTree{}
}

func (st *StoreTree) AddStore(s *Store) {
	st.Datacenter(s.GetLabel(LabelDatacenter)).
		Rack(s.GetLabel(LabelRack)).
		Host(s.GetLabel(LabelHost)).
		AddStore(s)
}

func (s StoreTree) String() string {
	buf := bytes.NewBuffer(nil)
	s.Pretty(0, buf)
	return buf.String()
}

func (s *StoreTree) Pretty(indent int, buf *bytes.Buffer) {
	buf.WriteString("_\n")
	for idx := range s.Datacenters {
		s.Datacenters[idx].Pretty(indent+1, buf)
	}
}

func (s *StoreTree) Datacenter(dcName string) *StoreTreeDatacenter {
	for idx := range s.Datacenters {
		if s.Datacenters[idx].Name == dcName {
			return &s.Datacenters[idx]
		}
	}
	s.Datacenters = append(s.Datacenters, StoreTreeDatacenter{
		Name: dcName,
	})
	return &s.Datacenters[len(s.Datacenters)-1]
}

type StoreTreeDatacenter struct {
	Name  string
	Racks []StoreTreeRack
}

func (s *StoreTreeDatacenter) Pretty(indent int, buf *bytes.Buffer) {
	printIndent(indent, buf)
	buf.WriteString(s.Name)
	buf.WriteString("\n")
	for idx := range s.Racks {
		s.Racks[idx].Pretty(indent+1, buf)
	}
}

func (s *StoreTreeDatacenter) Rack(rackName string) *StoreTreeRack {
	for idx := range s.Racks {
		if s.Racks[idx].Name == rackName {
			return &s.Racks[idx]
		}
	}
	s.Racks = append(s.Racks, StoreTreeRack{
		Name: rackName,
	})
	return &s.Racks[len(s.Racks)-1]
}

type StoreTreeRack struct {
	Name  string
	Hosts []StoreTreeHost
}

func (s *StoreTreeRack) Host(hostName string) *StoreTreeHost {
	for idx := range s.Hosts {
		if s.Hosts[idx].Name == hostName {
			return &s.Hosts[idx]
		}
	}
	s.Hosts = append(s.Hosts, StoreTreeHost{
		Name: hostName,
	})
	return &s.Hosts[len(s.Hosts)-1]
}

func printIndent(indent int, buf *bytes.Buffer) {
	if indent > 1 {
		buf.WriteString(strings.Repeat("|   ", indent-1))
	}
	buf.WriteString(strings.Repeat("|-- ", 1))
}

func (s *StoreTreeRack) Pretty(indent int, buf *bytes.Buffer) {
	printIndent(indent, buf)
	buf.WriteString(s.Name)
	buf.WriteString("\n")
	for idx := range s.Hosts {
		s.Hosts[idx].Pretty(indent+1, buf)
	}
}

type StoreTreeHost struct {
	Name   string
	Stores []*Store
}

func (s *StoreTreeHost) AddStore(store *Store) {
	s.Stores = append(s.Stores, store)
}

func (s *StoreTreeHost) Pretty(indent int, buf *bytes.Buffer) {
	printIndent(indent, buf)
	buf.WriteString(s.Name)
	buf.WriteString("\n")
	for idx := range s.Stores {
		printIndent(indent+1, buf)
		buf.WriteString(fmt.Sprintf("%v (%v)",
			s.Stores[idx].ID, s.Stores[idx].GetLabel(LabelStorageType)))
		buf.WriteString("\n")
	}
}
