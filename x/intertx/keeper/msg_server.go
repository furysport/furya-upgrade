package keeper

import (
	"context"
	"time"

	"cosmossdk.io/errors"
	"github.com/furysport/furya-upgrade/x/intertx/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmostypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the intertx Msg service interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterAccount implements the Msg/RegisterAccount interface
func (k msgServer) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgRegister := icacontrollertypes.NewMsgRegisterInterchainAccount(msg.ConnectionId, msg.Owner, "")
	icaMsgServer := icacontrollerkeeper.NewMsgServerImpl(&k.Keeper.icaControllerKeeper)

	if _, err := icaMsgServer.RegisterInterchainAccount(
		sdk.WrapSDKContext(ctx),
		msgRegister,
	); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAccountResponse{}, nil
}

// SubmitTx implements the Msg/SubmitTx interface
func (k msgServer) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	data, err := SerializeCosmosTx(k.Keeper.cdc, msg.Msgs)
	if err != nil {
		return nil, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	msgServer := icacontrollerkeeper.NewMsgServerImpl(&k.icaControllerKeeper)
	_, err = msgServer.SendTx(ctx, icacontrollertypes.NewMsgSendTx(msg.Owner, msg.ConnectionId, uint64(timeoutTimestamp), packetData))
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTxResponse{}, nil
}

func SerializeCosmosTx(cdc codec.BinaryCodec, msgs []*cosmostypes.Any) (bz []byte, err error) {
	// only ProtoCodec is supported
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, errors.Wrap(err, "only ProtoCodec is supported on host chain")
	}

	cosmosTx := &icatypes.CosmosTx{
		Messages: msgs,
	}

	bz, err = cdc.Marshal(cosmosTx)
	if err != nil {
		return nil, err
	}

	return bz, nil
}
