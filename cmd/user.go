// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/fanux/sealctl/user"
	"net"

	"github.com/spf13/cobra"
)

var conf user.Config
var ips []string

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Kubernetes multi tencent command line tool",
	Long: `Easy to use this to create a kubernetes user, 
           If your want some one access your kubernetes cluster read only, 
           you can use this command generate a kubeconfig for him, and bind 
           read only role etc..`,
	Run: func(cmd *cobra.Command, args []string) {
		for _,ip := range ips {
			conf.IPAddresses = append(conf.IPAddresses,net.ParseIP(ip))
		}
		err := user.GenerateKubeconfig(conf)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	userCmd.Flags().StringVarP(&conf.User,"user","u","fanux","user name in your kube config")
	userCmd.Flags().StringSliceVarP(&conf.Groups,"group","g",[]string{"sealyun","alibaba"},"user group names")
	userCmd.Flags().StringVarP(&conf.OutPut,"out","o","./kube/config","default kube config out put file name")
	userCmd.Flags().StringVar(&conf.CACrtFile,"ca-crt","/etc/kubernetes/ca.crt","kubernetes ca crt file")
	userCmd.Flags().StringVar(&conf.CAKeyFile,"ca-key","/etc/kubernetes/ca.key","kubernetes ca key file")
	userCmd.Flags().StringVar(&conf.ClusterName,"cluster-name","kubernetes","kubeconfig cluster name")
	userCmd.Flags().StringVarP(&conf.Apiserver,"apiserver","s","https://apiserver.cluster.local:6443","apiserver address")
	userCmd.Flags().StringSliceVarP(&conf.DNSNames,"dns","d",[]string{"apiserver.cluster.local", "localhost","sealyun.com"},"apiserver certSANs dns list")
	userCmd.Flags().StringSliceVar(&ips,"ips",[]string{"127.0.0.1","10.103.97.2"},"apiserver certSANs ip list")
}
