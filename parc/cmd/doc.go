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

	finalpath := path.Join(*outputDir, *outputFile)

	err := docs.GenerateDocumentation(cmd.Root(), finalpath)

	if err != nil {
		panic(err)
	}
}

var outputDir *string
var outputFile *string

func init() {
	rootCmd.AddCommand(genDocCmd)
	outputDir = genDocCmd.Flags().String("path", "", "Path to generate documentation files")
	outputFile = genDocCmd.Flags().String("filename", "doc.yaml", "Filename to use")
	genDocCmd.MarkFlagDirname("path")
	genDocCmd.MarkFlagFilename("filename")
}
