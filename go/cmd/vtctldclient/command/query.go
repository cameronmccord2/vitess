/*
Copyright 2022 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"vitess.io/vitess/go/cmd/vtctldclient/cli"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/topo/topoproto"

	vtctldatapb "vitess.io/vitess/go/vt/proto/vtctldata"
)

var (
	// ExecuteFetchAsApp makes an ExecuteFetchAsApp gRPC call to a vtctld.
	ExecuteFetchAsApp = &cobra.Command{
		Use:                   "ExecuteFetchAsApp [--max-rows <max-rows>] [--json|-j] [--use-pool] <tablet-alias> <query>",
		Short:                 "Executes the given query as the App user on the remote tablet.",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(2),
		RunE:                  commandExecuteFetchAsApp,
	}
	// ExecuteFetchAsDBA makes an ExecuteFetchAsDBA gRPC call to a vtctld.
	ExecuteFetchAsDBA = &cobra.Command{
		Use:                   "ExecuteFetchAsDBA [--max-rows <max-rows>] [--json|-j] [--disable-binlogs] [--reload-schema] <tablet alias> <query>",
		Short:                 "Executes the given query as the DBA user on the remote tablet.",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(2),
		RunE:                  commandExecuteFetchAsDBA,
		Aliases:               []string{"ExecuteFetchAsDba"},
	}
	// ExecuteMultiFetchAsDBA makes an ExecuteMultiFetchAsDBA gRPC call to a vtctld.
	ExecuteMultiFetchAsDBA = &cobra.Command{
		Use:                   "ExecuteMultiFetchAsDBA [--max-rows <max-rows>] [--json|-j] [--disable-binlogs] [--reload-schema] <tablet alias> <sql>",
		Short:                 "Executes given multiple queries as the DBA user on the remote tablet.",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(2),
		RunE:                  commandExecuteMultiFetchAsDBA,
		Aliases:               []string{"ExecuteMultiFetchAsDba"},
	}
	// GetUnresolvedTransactions makes an GetUnresolvedTransactions gRPC call to a vtctld.
	GetUnresolvedTransactions = &cobra.Command{
		Use:   "GetUnresolvedTransactions <keyspace>",
		Short: "Retrieves unresolved transactions for the given keyspace.",
		Args:  cobra.ExactArgs(1),
		RunE:  commandGetUnresolvedTransactions,
	}
)

var executeFetchAsAppOptions = struct {
	MaxRows int64
	UsePool bool
	JSON    bool
}{
	MaxRows: 10_000,
}

func commandExecuteFetchAsApp(cmd *cobra.Command, args []string) error {
	alias, err := topoproto.ParseTabletAlias(cmd.Flags().Arg(0))
	if err != nil {
		return err
	}

	cli.FinishedParsing(cmd)

	query := cmd.Flags().Arg(1)

	resp, err := client.ExecuteFetchAsApp(commandCtx, &vtctldatapb.ExecuteFetchAsAppRequest{
		TabletAlias: alias,
		Query:       query,
		MaxRows:     executeFetchAsAppOptions.MaxRows,
		UsePool:     executeFetchAsAppOptions.UsePool,
	})
	if err != nil {
		return err
	}

	qr := sqltypes.Proto3ToResult(resp.Result)
	switch executeFetchAsAppOptions.JSON {
	case true:
		data, err := cli.MarshalJSON(qr)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", data)
	default:
		cli.WriteQueryResultTable(cmd.OutOrStdout(), qr)
	}

	return nil
}

var executeFetchAsDBAOptions = struct {
	MaxRows        int64
	DisableBinlogs bool
	ReloadSchema   bool
	JSON           bool
}{
	MaxRows: 10_000,
}

func commandExecuteFetchAsDBA(cmd *cobra.Command, args []string) error {
	alias, err := topoproto.ParseTabletAlias(cmd.Flags().Arg(0))
	if err != nil {
		return err
	}

	cli.FinishedParsing(cmd)

	query := cmd.Flags().Arg(1)

	resp, err := client.ExecuteFetchAsDBA(commandCtx, &vtctldatapb.ExecuteFetchAsDBARequest{
		TabletAlias:    alias,
		Query:          query,
		MaxRows:        executeFetchAsDBAOptions.MaxRows,
		DisableBinlogs: executeFetchAsDBAOptions.DisableBinlogs,
		ReloadSchema:   executeFetchAsDBAOptions.ReloadSchema,
	})
	if err != nil {
		return err
	}

	qr := sqltypes.Proto3ToResult(resp.Result)
	switch executeFetchAsDBAOptions.JSON {
	case true:
		data, err := cli.MarshalJSON(qr)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", data)
	default:
		cli.WriteQueryResultTable(cmd.OutOrStdout(), qr)
	}

	return nil
}

var executeMultiFetchAsDBAOptions = struct {
	MaxRows        int64
	DisableBinlogs bool
	ReloadSchema   bool
	JSON           bool
}{
	MaxRows: 10_000,
}

func commandExecuteMultiFetchAsDBA(cmd *cobra.Command, args []string) error {
	alias, err := topoproto.ParseTabletAlias(cmd.Flags().Arg(0))
	if err != nil {
		return err
	}

	cli.FinishedParsing(cmd)

	sql := cmd.Flags().Arg(1)

	resp, err := client.ExecuteMultiFetchAsDBA(commandCtx, &vtctldatapb.ExecuteMultiFetchAsDBARequest{
		TabletAlias:    alias,
		Sql:            sql,
		MaxRows:        executeMultiFetchAsDBAOptions.MaxRows,
		DisableBinlogs: executeMultiFetchAsDBAOptions.DisableBinlogs,
		ReloadSchema:   executeMultiFetchAsDBAOptions.ReloadSchema,
	})
	if err != nil {
		return err
	}

	var qrs []*sqltypes.Result
	for _, result := range resp.Results {
		qr := sqltypes.Proto3ToResult(result)
		qrs = append(qrs, qr)
	}

	switch executeMultiFetchAsDBAOptions.JSON {
	case true:
		data, err := cli.MarshalJSON(qrs)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", data)
	default:
		for _, qr := range qrs {
			cli.WriteQueryResultTable(cmd.OutOrStdout(), qr)
		}
	}
	return nil
}

func commandGetUnresolvedTransactions(cmd *cobra.Command, args []string) error {
	cli.FinishedParsing(cmd)

	keyspace := cmd.Flags().Arg(0)
	resp, err := client.GetUnresolvedTransactions(commandCtx,
		&vtctldatapb.GetUnresolvedTransactionsRequest{
			Keyspace: keyspace,
		})
	if err != nil {
		return err
	}

	data, err := cli.MarshalJSON(resp.Transactions)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", data)
	return nil
}

func init() {
	ExecuteFetchAsApp.Flags().Int64Var(&executeFetchAsAppOptions.MaxRows, "max-rows", 10_000, "The maximum number of rows to fetch from the remote tablet.")
	ExecuteFetchAsApp.Flags().BoolVar(&executeFetchAsAppOptions.UsePool, "use-pool", false, "Use the tablet connection pool instead of creating a fresh connection.")
	ExecuteFetchAsApp.Flags().BoolVarP(&executeFetchAsAppOptions.JSON, "json", "j", false, "Output the results in JSON instead of a human-readable table.")
	Root.AddCommand(ExecuteFetchAsApp)

	ExecuteFetchAsDBA.Flags().Int64Var(&executeFetchAsDBAOptions.MaxRows, "max-rows", 10_000, "The maximum number of rows to fetch from the remote tablet.")
	ExecuteFetchAsDBA.Flags().BoolVar(&executeFetchAsDBAOptions.DisableBinlogs, "disable-binlogs", false, "Disables binary logging during the query.")
	ExecuteFetchAsDBA.Flags().BoolVar(&executeFetchAsDBAOptions.ReloadSchema, "reload-schema", false, "Instructs the tablet to reload its schema after executing the query.")
	ExecuteFetchAsDBA.Flags().BoolVarP(&executeFetchAsDBAOptions.JSON, "json", "j", false, "Output the results in JSON instead of a human-readable table.")
	Root.AddCommand(ExecuteFetchAsDBA)

	ExecuteMultiFetchAsDBA.Flags().Int64Var(&executeMultiFetchAsDBAOptions.MaxRows, "max-rows", 10_000, "The maximum number of rows to fetch from the remote tablet.")
	ExecuteMultiFetchAsDBA.Flags().BoolVar(&executeMultiFetchAsDBAOptions.DisableBinlogs, "disable-binlogs", false, "Disables binary logging during the query.")
	ExecuteMultiFetchAsDBA.Flags().BoolVar(&executeMultiFetchAsDBAOptions.ReloadSchema, "reload-schema", false, "Instructs the tablet to reload its schema after executing the query.")
	ExecuteMultiFetchAsDBA.Flags().BoolVarP(&executeMultiFetchAsDBAOptions.JSON, "json", "j", false, "Output the results in JSON instead of a human-readable table.")
	Root.AddCommand(ExecuteMultiFetchAsDBA)

	Root.AddCommand(GetUnresolvedTransactions)
}
