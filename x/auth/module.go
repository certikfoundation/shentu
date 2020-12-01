package auth

import (
	"encoding/json"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authSim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	sim "github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/certikfoundation/shentu/x/auth/client/cli"
	"github.com/certikfoundation/shentu/x/auth/internal/keeper"
	"github.com/certikfoundation/shentu/x/auth/simulation"
	"github.com/certikfoundation/shentu/x/auth/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the auth module.
type AppModuleBasic struct {
	cdc codec.Marshaler
}

// Name returns the auth module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the auth module's types for the given codec.
//func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
//	RegisterCodec(cdc)
//	*CosmosModuleCdc = *ModuleCdc // nolint
//}

// DefaultGenesis returns default genesis state as raw bytes for the auth module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return CosmosAppModuleBasic{}.DefaultGenesis(cdc)
}

// ValidateGenesis performs genesis state validation for the auth module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	return CosmosAppModuleBasic{}.ValidateGenesis(cdc, config, bz)
}

// RegisterRESTRoutes registers the REST routes for the auth module.
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {
	RegisterRoutes(ctx, rtr)
	CosmosAppModuleBasic{}.RegisterRESTRoutes(ctx, rtr)
}

// TODO
// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the staking module.
//func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
//	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
//}

// GetTxCmd returns the root tx command for the auth module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns the root query command for the auth module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return CosmosAppModuleBasic{}.GetQueryCmd()
}

//____________________________________________________________________________

// AppModule implements an application module for the auth module.
type AppModule struct {
	AppModuleBasic
	cosmosAppModule CosmosAppModule

	keeper     keeper.Keeper
	authKeeper AccountKeeper
	certKeeper types.CertKeeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Marshaler, keeper keeper.Keeper, ak AccountKeeper, ck types.CertKeeper) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{cdc: cdc},
		cosmosAppModule: NewCosmosAppModule(cdc, ak),
		keeper:          keeper,
		authKeeper:      ak,
		certKeeper:      ck,
	}
}

// Name returns the auth module's name.
func (am AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants performs a no-op.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	am.cosmosAppModule.RegisterInvariants(ir)
}

// Route returns the message routing key for the auth module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

// NewHandler returns an sdk.Handler for the auth module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.authKeeper)
}

// QuerierRoute returns the auth module's querier route name.
func (am AppModule) QuerierRoute() string {
	return am.cosmosAppModule.QuerierRoute()
}

// NewQuerierHandler returns the auth module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return am.cosmosAppModule.NewQuerierHandler()
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	querier := keeper.Querier{Keeper: am.keeper}
	types.RegisterQueryServer(cfg.QueryServer(), querier)
}

// InitGenesis performs genesis initialization for the auth module. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return am.cosmosAppModule.InitGenesis(ctx, cdc, data)
}

// ExportGenesis returns the exported genesis state as raw bytes for the auth module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return am.cosmosAppModule.ExportGenesis(ctx, cdc)
}

// BeginBlock returns the begin blocker for the auth module.
func (am AppModule) BeginBlock(ctx sdk.Context, rbb abci.RequestBeginBlock) {
	am.cosmosAppModule.BeginBlock(ctx, rbb)
}

// EndBlock returns the end blocker for the auth module. It returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, rbb abci.RequestEndBlock) []abci.ValidatorUpdate {
	return am.cosmosAppModule.EndBlock(ctx, rbb)
}

//____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the auth module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized auth param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []sim.ParamChange {
	return authSim.ParamChanges(r)
}

// RegisterStoreDecoder registers a decoder for auth module's types.
func (AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[StoreKey] = authSim.DecodeStore
}

// WeightedOperations returns auth operations for use in simulations.
func (am AppModule) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.authKeeper)
}
