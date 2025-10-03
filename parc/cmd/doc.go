package cmd

import (
	"github.com/spf13/cobra"
)

var genDocCmd = &cobra.Command{
	Use:   "Generate documentation for web",
	Short: "Generate docs",
	Long:  "Generate Markdown/manual pages for the application",
	Run:   genDocsMain,
}

func docMakeFrontmatter(k string) string {
	return ""
}

func docMakeLink(n string) string {
	return n
}

func genDocsMain(cmd *cobra.Command, args []string) {

	for _, c := range rootCmd.Commands() {
		generatePage(c)
	}

}

func generatePage(c *cobra.Command) {
	//
	panic("xxxx")
}

var outputDir *string

func init() {
	rootCmd.AddCommand(genDocCmd)
	outputDir = genDocCmd.Flags().String("path", "./docs/parc/", "Path to generate documentation files")
}
