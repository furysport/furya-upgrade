package furya

import (
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/furysport/furya-upgrade/app/keepers"
	"github.com/furysport/furya-upgrade/app/upgrades"
	v130 "github.com/furysport/furya-upgrade/app/upgrades/v130"
	v131 "github.com/furysport/furya-upgrade/app/upgrades/v131"
	v140 "github.com/furysport/furya-upgrade/app/upgrades/v140"
	v200 "github.com/furysport/furya-upgrade/app/upgrades/v200"
	airdrop "github.com/furysport/furya-upgrade/x/airdrop"
	airdropkeeper "github.com/furysport/furya-upgrade/x/airdrop/keeper"
	airdroptypes "github.com/furysport/furya-upgrade/x/airdrop/types"
	"github.com/furysport/furya-upgrade/x/mint"
	mintkeeper "github.com/furysport/furya-upgrade/x/mint/keeper"
	minttypes "github.com/furysport/furya-upgrade/x/mint/types"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	intertx "github.com/furysport/furya-upgrade/x/intertx"
	intertxkeeper "github.com/furysport/furya-upgrade/x/intertx/keeper"
	intertxtypes "github.com/furysport/furya-upgrade/x/intertx/types"
	
	bank "github.com/terra-money/alliance/custom/bank"
	custombankkeeper "github.com/terra-money/alliance/custom/bank/keeper"

	furyaappparams "github.com/furysport/furya-upgrade/app/params"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/prometheus/client_golang/prometheus"
	
	alliancemodule "github.com/terra-money/alliance/x/alliance"
	alliancemoduleclient "github.com/terra-money/alliance/x/alliance/client"
	alliancemodulekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	alliancemoduletypes "github.com/terra-money/alliance/x/alliance/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/cosmos/cosmos-sdk/client/docs/statik"
)

const appName = "FuryaApp"

var (
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	ProposalsEnabled = "true"
	// If set to non-empty string it must be comma-separated list of values that are all a subset
	// of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificProposals = ""
)

// GetEnabledProposals parses the ProposalsEnabled / EnableSpecificProposals values to
// produce a list of enabled proposals to pass into wasmd app.
func GetEnabledProposals() []wasmtypes.ProposalType {
	if EnableSpecificProposals == "" {
		if ProposalsEnabled == "true" {
			return wasmtypes.EnableAllProposals
		}
		return wasmtypes.DisableAllProposals
	}
	chunks := strings.Split(EnableSpecificProposals, ",")
	proposals, err := wasmtypes.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}
	return proposals
}

func GetWasmOpts(appOpts servertypes.AppOptions) []wasmkeeper.Option {
	var wasmOpts []wasmkeeper.Option
	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	wasmOpts = append(wasmOpts, wasmkeeper.WithGasRegister(NewPoppWasmGasRegister()))

	return wasmOpts
}

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		alliancemoduleclient.CreateAllianceProposalHandler,
		alliancemoduleclient.UpdateAllianceProposalHandler,
		alliancemoduleclient.DeleteAllianceProposalHandler,
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	Upgrades = []upgrades.Upgrade{v130.Upgrade, v131.Upgrade, v140.Upgrade, v200.Upgrade}

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		airdrop.AppModuleBasic{},
		auth.AppModuleBasic{},
		alliancemodule.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModule{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		consensus.AppModuleBasic{},
		packetforward.AppModuleBasic{},
		wasm.AppModuleBasic{},
		ibcfee.AppModuleBasic{},
		ica.AppModuleBasic{},
		intertx.AppModule{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     					nil,
		distrtypes.ModuleName:          					nil,
		icatypes.ModuleName:            					nil,
		minttypes.ModuleName:           					{authtypes.Minter},
		stakingtypes.BondedPoolName:    				{authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: 				{authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            					{authtypes.Burner},
		ibctransfertypes.ModuleName:    					{authtypes.Minter, authtypes.Burner},
		airdroptypes.ModuleName:        					nil,
		ibcfeetypes.ModuleName:         					nil,
		wasmtypes.ModuleName:           				{authtypes.Burner},
		alliancemoduletypes.ModuleName:      		{authtypes.Minter, authtypes.Burner},
		alliancemoduletypes.RewardsPoolName: 		nil,
	}
)

var (
	_ runtime.AppI            				= (*FuryaApp)(nil)
	_ servertypes.Application 			= (*FuryaApp)(nil)
)

// FuryaApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type FuryaApp struct { // nolint: golint
	*baseapp.BaseApp
	keepers.AppKeepers

	legacyAmino       		*codec.LegacyAmino
	appCodec          		codec.Codec
	interfaceRegistry 		types.InterfaceRegistry

	// keys to access the substores
	keys    			map[string]*storetypes.KVStoreKey
	tkeys   			map[string]*storetypes.TransientStoreKey
	memKeys 		map[string]*storetypes.MemoryStoreKey

	// the module manager
	mm 			*module.Manager

	// simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		stdlog.Println("Failed to get home dir %2", err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".furyad")
}

// NewFuryaApp returns a reference to an initialized NxtPop.
func NewFuryaApp(
	logger log.Logger,
	db dbm.DB, traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	encodingConfig furyaappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *FuryaApp {

	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibcexported.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey, feegrant.StoreKey, authzkeeper.StoreKey, packetforwardtypes.StoreKey,
		icacontrollertypes.StoreKey, group.StoreKey,
		alliancemoduletypes.StoreKey,
		ibcfeetypes.StoreKey,
		icahosttypes.StoreKey,
		airdroptypes.StoreKey,
		wasmtypes.StoreKey,
		consensusparamtypes.StoreKey,
		intertxtypes.StoreKey,
		crisistypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &FuryaApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec, keys[consensusparamtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedInterTxKeeper := app.CapabilityKeeper.ScopeToModule(intertxtypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)

	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32PrefixAccAddr,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.BankKeeper = custombankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)
	groupConfig := group.DefaultConfig()
	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.MsgServiceRouter(), app.AccountKeeper, groupConfig)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		app.GetSubspace(minttypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		stakingKeeper,
		app.DistrKeeper,
		authtypes.FeeCollectorName,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	
	app.AllianceKeeper = alliancemodulekeeper.NewKeeper(
		appCodec,
		keys[alliancemoduletypes.StoreKey],
		app.GetSubspace(alliancemoduletypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		&stakingKeeper,
		app.DistrKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(), app.AllianceKeeper.StakingHooks()),
	)
	
	app.BankKeeper.RegisterKeepers(app.AllianceKeeper, &stakingKeeper)

	app.AirdropKeeper = *airdropkeeper.NewKeeper(appCodec, keys[airdroptypes.StoreKey], app.GetSubspace(airdroptypes.ModuleName), app.BankKeeper, app.StakingKeeper, app.AccountKeeper)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)
	app.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		appCodec,
		keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)

	govConfig := govtypes.DefaultConfig()
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// register the proposal types
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(alliancemoduletypes.RouterKey, alliancemodule.NewAllianceProposalHandler(app.AllianceKeeper))

	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey], 
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, app.MsgServiceRouter(),
	)
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)

	app.InterTxKeeper = intertxkeeper.NewKeeper(appCodec, keys[intertxtypes.StoreKey], app.ICAControllerKeeper, scopedInterTxKeeper)
	interTxModule := intertx.NewAppModule(appCodec, app.InterTxKeeper)

	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	var icaControllerStack ibcporttypes.IBCModule
	icaControllerStack = intertx.NewIBCModule(app.InterTxKeeper)
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, app.ICAControllerKeeper)
	icaControllerStack = ibcfee.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper)

	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		app.TransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.DistrKeeper,
		app.BankKeeper,
		app.IBCFeeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	packetforwardModule := packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName))

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)

	app.EvidenceKeeper = *evidenceKeeper

	wasmDir := filepath.Join(homePath, "data")

	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	app.setupUpgradeStoreLoaders()

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	availableCapabilities := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2"
	wasmOpts := GetWasmOpts(appOpts)
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		nil,
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// register wasm gov proposal types
	enabledProposals := GetEnabledProposals()
	if len(enabledProposals) != 0 {
		govRouter.AddRoute(wasmtypes.RouterKey, wasmkeeper.NewWasmProposalHandler(app.WasmKeeper, enabledProposals)) //nolint:staticcheck
	}
	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	var wasmStack ibcporttypes.IBCModule
	wasmStack = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.IBCFeeKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule).
		AddRoute(intertxtypes.ModuleName, icaControllerStack).
		AddRoute(wasmtypes.ModuleName, wasmStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		airdrop.NewAppModule(appCodec, app.AirdropKeeper),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		custombankmodule.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)), // always be last to make sure that it checks for all invariants and not only part of them
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, app.BankKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)), staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		transferModule,
		icaModule,
		alliancemodule.NewAppModule(appCodec, app.AllianceKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		packetforwardModule,
		interTxModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		crisistypes.ModuleName,
		authtypes.ModuleName,
		paramstypes.ModuleName,
		govtypes.ModuleName,
		feegrant.ModuleName,
		ibctransfertypes.ModuleName,
		genutiltypes.ModuleName,
		banktypes.ModuleName,
		authz.ModuleName,
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		group.ModuleName,
		icatypes.ModuleName,
		packetforwardtypes.ModuleName,
		airdroptypes.ModuleName,
		wasmtypes.ModuleName,
		intertxtypes.ModuleName,
		alliancemoduletypes.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		distrtypes.ModuleName,
		ibctransfertypes.ModuleName,
		minttypes.ModuleName,
		upgradetypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		capabilitytypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		ibcexported.ModuleName,
		group.ModuleName,
		icatypes.ModuleName,
		packetforwardtypes.ModuleName,
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		airdroptypes.ModuleName,
		wasmtypes.ModuleName,
		intertxtypes.ModuleName,
		alliancemoduletypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		upgradetypes.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		authz.ModuleName,
		packetforwardtypes.ModuleName,
		airdroptypes.ModuleName,
		wasmtypes.ModuleName,
		intertxtypes.ModuleName,
		alliancemoduletypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.setupUpgradeHandlers()

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   		app.AccountKeeper,
				BankKeeper:      		app.BankKeeper,
				FeegrantKeeper:  		app.FeeGrantKeeper,
				SignModeHandler: 		encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  		ante.DefaultSigVerificationGasConsumer,
			},
			StakingKeeper:     			app.StakingKeeper,
			BankKeeper:        			app.BankKeeper,
			AirdropKeeper:     			&app.AirdropKeeper,
			IBCKeeper:         			app.IBCKeeper,
			TxCounterStoreKey: 		keys[wasmtypes.StoreKey],
			WasmConfig:        			wasmConfig,
			Cdc:               				appCodec,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// Ensure that state sync delivers the cosmwasm directory as well

	if manager := app.SnapshotManager(); manager != nil {
		err = manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic("failed to register snapshot extension: " + err.Error())
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}

		// Initialize and seal the capability keeper so all persistent capabilities
		// are loaded in-memory and prevent any further modules from creating scoped
		// sub-keepers.
		// This must be done during creation of baseapp rather than in InitChain so
		// that in-memory capabilities get regenerated on app restart.
		// Note that since this reads from the store, we can only perform it when
		// `loadLatest` is set to true.
		app.CapabilityKeeper.Seal()
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedInterTxKeeper = scopedInterTxKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper

	return app
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	delete(modAccAddrs, authtypes.NewModuleAddress(alliancemoduletypes.ModuleName).String())
	delete(modAccAddrs, authtypes.NewModuleAddress(authtypes.FeeCollectorName).String())

	return modAccAddrs
}
// Name returns the name of the App
func (app *FuryaApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *FuryaApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *FuryaApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *FuryaApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *FuryaApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *FuryaApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	return modAccAddrs
}


// LegacyAmino returns FuryaApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *FuryaApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns NxtPop's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *FuryaApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns NxtPop's InterfaceRegistry
func (app *FuryaApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *FuryaApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *FuryaApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *FuryaApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *FuryaApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *FuryaApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *FuryaApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *FuryaApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *FuryaApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *FuryaApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// GetBaseApp implements the BaseApp
func (app *FuryaApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetBaseApp implements the StakingKeeper
func (app *FuryaApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetBaseApp implements the StakingKeeper
func (app *FuryaApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetBaseApp implements the StakingKeeper
func (app *FuryaApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetBaseApp implements the StakingKeeper
func (app *FuryaApp) GetTxConfig() client.TxConfig {
	return MakeEncodingConfig().TxConfig
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(airdroptypes.ModuleName)
	paramsKeeper.Subspace(alliancemoduletypes.ModuleName)

	return paramsKeeper
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *FuryaApp) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
		}
	}
}

func (app *FuryaApp) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				app.AppKeepers,
			),
		)
	}
}
