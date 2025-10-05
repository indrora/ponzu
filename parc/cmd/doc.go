package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/indrora/ponzu/parc/docs"
	"github.com/spf13/cobra"
)

var genDocCmd = &cobra.Command{
	Use:   "gendocs",
	Short: "Generate docs",
	Long:  "Generate Markdown/manual pages for the application",
	Run:   genDocsMain,
}

func docMakeLink(n string) string {
	return fmt.Sprintf("[ %s ]({{< ref %s >}})", n, strings.ReplaceAll(n, "_", "-"))
}

func genDocsMain(cmd *cobra.Command, args []string) {

	finalpath := path.Join(*outputDir, "doc.yaml")

	err := docs.GenerateDocumentation(cmd.Root(), finalpath)

	if err != nil {
		panic(err)
	}
}

var outputDir *string

func init() {
	rootCmd.AddCommand(genDocCmd)
	outputDir = genDocCmd.Flags().String("path", "./docs/data/", "Path to generate documentation files")
}
