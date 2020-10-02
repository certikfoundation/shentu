package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/certikfoundation/shentu/x/shield/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	shieldQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the shield module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	shieldQueryCmd.AddCommand(flags.GetCommands(
		GetCmdPool(queryRoute, cdc),
		GetCmdPurchase(queryRoute, cdc),
		GetCmdProvider(queryRoute, cdc),
	)...)

	return shieldQueryCmd
}

// GetCmdPool returns the command for querying the pool.
func GetCmdPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Short: "get pool information",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			var res []byte
			var err error
			if len(args) == 1 {
				res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pool/id/%s", queryRoute, args[0]), nil)
				if err != nil {
					return err
				}
			} else {
				sponsor := viper.GetString(flagSponsor)
				if sponsor == "" {
					return fmt.Errorf("either poolID or sponsor is required to query pool")
				}
				res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pool/sponsor/%s", queryRoute, sponsor), nil)
				if err != nil {
					return err
				}
			}
			var out types.Pool
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
	cmd.Flags().String(flagSponsor, "", "use sponsor to query the pool info")

	return cmd
}

// GetCmdPurchase returns the command for querying a purchase.
func GetCmdPurchase(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purchase [txhash]",
		Short: "get purchase information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/purchase/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			var out types.Purchase
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}

// GetCmdProvider returns the command for querying a provider.
func GetCmdProvider(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider [address]",
		Short: "get provider information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/provider/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			var out types.Provider
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	return cmd
}