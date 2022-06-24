package miner

import (
	"github.com/abeychain/go-abey/abeydb"
	"github.com/abeychain/go-abey/accounts"
	"github.com/abeychain/go-abey/cmd/utils"
	"github.com/abeychain/go-abey/common"
	"github.com/abeychain/go-abey/consensus"
	"github.com/abeychain/go-abey/consensus/minerva"
	"github.com/abeychain/go-abey/core"
	"github.com/abeychain/go-abey/core/snailchain"
	"github.com/abeychain/go-abey/core/types"
	"github.com/abeychain/go-abey/core/vm"
	"github.com/abeychain/go-abey/params"
	"testing"
)

type mockBackend struct {
	db             abeydb.Database
	txPool         *core.TxPool
	schain         *snailchain.SnailBlockChain
	fchain         *core.BlockChain
	uncleBlock     *types.Block
	snailPool      *snailchain.SnailPool
	accountManager *accounts.Manager
}

func newMockBackend(fastchaincfg *params.ChainConfig, engine consensus.Engine) *mockBackend {
	var (
		db      = abeydb.NewMemDatabase()
		genesis = core.DefaultDevGenesisBlock()
		cache   = &core.CacheConfig{}
		vmcfg   = vm.Config{}
		//fastchaincfg = params.DevnetChainConfig
		//engine       = minerva.NewFaker()
		fastNums = 10 * params.MinimumFruits
	)
	// make fast chain
	fchain, err := core.NewBlockChain(db, cache, fastchaincfg, engine, vmcfg)
	if err != nil {
		utils.Fatalf("failed to make new fast chain %v", err)
	}
	// make fast blocks
	fastGenesis := genesis.MustFastCommit(db)
	fastblocks, _ := core.GenerateChain(params.TestChainConfig, fastGenesis, engine, db, fastNums, func(i int, b *core.BlockGen) {
		b.SetCoinbase(common.Address{0: byte(1), 19: byte(i)})
	})
	fchain.InsertChain(fastblocks)

	// make the snail chain
	snailGenesis := genesis.MustSnailCommit(db)
	schain, err := snailchain.NewSnailBlockChain(db, fastchaincfg, engine, fchain)
	if err != nil {
		utils.Fatalf("failed to make new snail chain %v", err)
	}
	if _, err := schain.InsertChain(types.SnailBlocks{snailGenesis}); err != nil {
		utils.Fatalf("failed to insert genesis block %v", err)
	}
	//_, err := MakeSnailBlockBlockChain(snailChain, fastchain, snailGenesis, snailBlockNumbers, 1)
	//if err != nil {
	//	utils.Fatalf("failed to make new snail blocks %v", err)
	//}
	return &mockBackend{
		db:        db,
		schain:    schain,
		fchain:    fchain,
		snailPool: snailchain.NewSnailPool(snailchain.DefaultSnailPoolConfig, fchain, schain, engine),
	}
}
func (b *mockBackend) SnailBlockChain() *snailchain.SnailBlockChain { return b.schain }
func (b *mockBackend) AccountManager() *accounts.Manager            { return b.accountManager }
func (b *mockBackend) SnailGenesis() *types.SnailBlock              { return b.schain.GetBlockByNumber(0) }
func (b *mockBackend) TxPool() *core.TxPool                         { return b.txPool }
func (b *mockBackend) BlockChain() *core.BlockChain                 { return b.fchain }
func (b *mockBackend) ChainDb() abeydb.Database                     { return b.db }
func (b *mockBackend) SnailPool() *snailchain.SnailPool             { return b.snailPool }

func TestMakeSnailBlock(t *testing.T) {
	// make
	var (
		fastchaincfg = params.DevnetChainConfig
		engine       = minerva.NewFaker()
	)

	backend := newMockBackend(fastchaincfg, engine)
	worker := newWorker(fastchaincfg, engine, coinbase, backend, nil)

	worker.commitNewWork()
}