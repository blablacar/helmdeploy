package main

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

var (
	tillerService   string
	tillerNamespace string
	rootCmd         = &cobra.Command{}
	deployCmd       = &cobra.Command{
		Use:  "deploy <release_manifest>",
		RunE: deploy,
		Args: cobra.ExactArgs(1),
	}
	diffCmd = &cobra.Command{
		Use:  "diff <release_manifest>",
		RunE: diff,
		Args: cobra.ExactArgs(1),
	}
	statusCmd = &cobra.Command{
		Use:  "status <release_manifest>",
		RunE: status,
		Args: cobra.ExactArgs(1),
	}
	helmHostCmd = &cobra.Command{
		Use:  "helm-host <release_manifest>",
		RunE: helmHost,
		Args: cobra.ExactArgs(1),
	}
	templateCmd = &cobra.Command{
		Use:  "template <release_manifest>",
		RunE: template,
		Args: cobra.ExactArgs(1),
	}
	lintCmd = &cobra.Command{
		Use:  "lint <release_manifest>",
		RunE: lint,
		Args: cobra.ExactArgs(1),
	}
)

func main() {

	log.SetLevel(log.DebugLevel)

	rootCmd.PersistentFlags().StringVar(&tillerNamespace, "tiller-namespace", "kube-system", "Tiller namespace")
	rootCmd.PersistentFlags().StringVar(&tillerService, "tiller-service", "tiller-deploy", "Tiller service name")

	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(helmHostCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(lintCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func deploy(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]

	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}

	hapiRelease, err := helmDeployer.Deploy(false)
	if err != nil {
		return err
	}

	if err := hapiRelease.PrintStatus(os.Stdout); err != nil {
		return err
	}

	return nil
}

func status(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]
	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}

	hapiRelease, err := helmDeployer.Status()
	if err != nil {
		return err
	}

	if err := hapiRelease.PrintStatus(os.Stdout); err != nil {
		return err
	}

	return nil
}

func diff(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]
	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}

	newRelease, err := helmDeployer.Deploy(true)
	if err != nil {
		return err
	}

	oldRelease, err := helmDeployer.Content()
	if err != nil {
		return err
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldRelease.Manifest),
		B:        difflib.SplitLines(newRelease.Manifest),
		FromFile: "",
		ToFile:   "",
		Context:  3,
	})
	fmt.Printf(diff)
	return nil
}

func helmHost(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]
	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}

	endpoints, err := helmDeployer.KubeClient.(*Clientset).GetEndpoints(tillerNamespace, tillerService)
	if err != nil {
		return err
	}

	if len(endpoints) > 0 {
		fmt.Printf("HELM_HOST=\"%s\"\n", endpoints[0])
	} else {
		return fmt.Errorf("Could not find any %s endpoint in namespace %s", tillerService, tillerNamespace)
	}
	return nil
}

func template(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]

	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}
	out, err := helmDeployer.Render()
	if err != nil {
		return err
	}
	for k, v := range out {
		fmt.Printf("# Source: %s\n", k)
		fmt.Println(v)
	}

	//
	return nil
}

func lint(cmd *cobra.Command, args []string) error {
	manifestPath := args[0]

	helmDeployer, err := NewDeployerFromManifest(manifestPath, tillerNamespace, tillerService)
	if err != nil {
		return err
	}
	out, err := helmDeployer.Render()

	if err != nil {
		return err
	}

	//TODO : for now we lint the k8s resource, not the helm chart.
	//tmpDir, err := ioutil.TempDir("", "helmdeploy-linter")
	//if err != nil {
	//	return err
	//}

	for resource, content := range out {
		fmt.Printf("# Source: %s\n", resource)
		err := LintResource(content)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("OK")
		}
	}
	return nil
}
