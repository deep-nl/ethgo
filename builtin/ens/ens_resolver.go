package ens

import (
	"github.com/deep-nl/ethgo/contract"
	"github.com/deep-nl/ethgo/core"
	"github.com/deep-nl/ethgo/jsonrpc"
	"strings"
)

type ENSResolver struct {
	e        *ENS
	provider *jsonrpc.Eth
}

func NewENSResolver(addr core.Address, provider *jsonrpc.Client) *ENSResolver {
	return &ENSResolver{NewENS(addr, contract.WithJsonRPC(provider.Eth())), provider.Eth()}
}

func (e *ENSResolver) Resolve(addr string, block ...core.BlockNumber) (res core.Address, err error) {
	addrHash := NameHash(addr)
	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Addr(addrHash, block...)
	return
}

func addressToReverseDomain(addr core.Address) string {
	return strings.ToLower(strings.TrimPrefix(addr.String(), "0x")) + ".addr.reverse"
}

func (e *ENSResolver) ReverseResolve(addr core.Address, block ...core.BlockNumber) (res string, err error) {
	addrHash := NameHash(addressToReverseDomain(addr))

	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Name(addrHash, block...)
	return
}

func (e *ENSResolver) prepareResolver(inputHash core.Hash, block ...core.BlockNumber) (*Resolver, error) {
	resolverAddr, err := e.e.Resolver(inputHash, block...)
	if err != nil {
		return nil, err
	}

	resolver := NewResolver(resolverAddr, contract.WithJsonRPC(e.provider))
	return resolver, nil
}
