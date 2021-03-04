// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package coretypes

import (
	"bytes"
	"errors"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/mr-tron/base58"
)

// AgentIDLength is the size of AgentID in bytes
const AgentIDLength = ContractIDLength // max(ContractIDLength, ledgerstate.AddressLength)

// AgentID represents exactly one of two types of entities on the ISCP ledger in one ID:
//  - It can represent an address on the Tangle (controlled by some private key). In this case it can be
//    interpreted as ledgerstate.Address type (see MustAddress).
//  - alternatively, it can represent a smart contract on the ISCP. In this case it can be interpreted as
//    a coretypes.ContractID type (see MustContractID)
// Type of ID represented by the AgentID can be recognized with IsAddress call.
// An attempt to interpret the AgentID in the wrong way invokes panic
type AgentID [AgentIDLength]byte

// NewAgentIDFromContractID makes AgentID from ContractID
func NewAgentIDFromContractID(id ContractID) (ret AgentID) {
	copy(ret[:], id[:])
	return
}

// NewAgentIDFromAddress makes AgentID from ledgerstate.Address
func NewAgentIDFromAddress(addr ledgerstate.Address) AgentID {
	// 0 is a reserved hname
	return NewAgentIDFromContractID(NewContractID(ChainID(addr.Array()), 0))
}

// NewAgentIDFromBytes makes an AgentID from binary representation
func NewAgentIDFromBytes(data []byte) (ret AgentID, err error) {
	if len(data) != AgentIDLength {
		err = ErrWrongDataLength
		return
	}
	copy(ret[:], data)
	return
}

func (a *AgentID) chainIDField() []byte {
	return a[:ChainIDLength]
}

func (a *AgentID) hnameField() []byte {
	return a[ChainIDLength : ChainIDLength+HnameLength]
}

// IsAddress checks if agentID represents address. 0 in the place of the contract's hname means it is an address
// This is based on the assumption that fro coretypes.Hname 0 is a reserved value
func (a AgentID) IsAddress() bool {
	var z [4]byte
	return bytes.Equal(a.hnameField(), z[:])
}

// MustAddress takes address or panic if not address
func (a AgentID) MustAddress() ledgerstate.Address {
	if !a.IsAddress() {
		panic("not an address")
	}
	addr, _, err := ledgerstate.AddressFromBytes(a.chainIDField())
	if err != nil {
		panic(err)
	}
	return addr
}

// MustContractID takes contract ID or panics if not a contract ID
func (a AgentID) MustContractID() (ret ContractID) {
	if a.IsAddress() {
		panic("not a contract")
	}
	copy(ret[:], a[:])
	return
}

// String human readable string
func (a AgentID) String() string {
	if a.IsAddress() {
		return "A/" + a.MustAddress().String()
	}
	cid := a.MustContractID()
	return "C/" + cid.String()
}

// NewAgentIDFromString parses the human-readable string representation
func NewAgentIDFromString(s string) (ret AgentID, err error) {
	if len(s) < 2 {
		err = errors.New("invalid length")
		return
	}
	switch s[:2] {
	case "A/":
		var addr ledgerstate.Address
		addr, err = ledgerstate.AddressFromBase58EncodedString(s[2:])
		if err != nil {
			return
		}
		return NewAgentIDFromAddress(addr), nil
	case "C/":
		var cid ContractID
		cid, err = NewContractIDFromString(s[2:])
		if err != nil {
			return
		}
		return NewAgentIDFromContractID(cid), nil
	default:
		err = errors.New("invalid prefix")
	}
	return
}

// ReadAgentID decodes from binary representation
func ReadAgentID(r io.Reader, agentID *AgentID) error {
	n, err := r.Read(agentID[:])
	if err != nil {
		return err
	}
	if n != AgentIDLength {
		return errors.New("error while reading agent ID")
	}
	return nil
}

func (a AgentID) Base58() string {
	return base58.Encode(a[:])
}
