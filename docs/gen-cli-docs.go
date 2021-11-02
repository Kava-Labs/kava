package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/cmd/kava/cmd"
)

func main() {
	root := cmd.NewRootCmd("")

	if err := GenMarkdownTreeCustom(root, "./docs/cli"); err != nil {
		log.Fatal(err)
	}

}

func GenMarkdownTreeCustom(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir); err != nil {
			return err
		}
	}

	hasChildren := len(cmd.Commands()) > 0

	cmdPath := strings.Split(cmd.CommandPath(), " ")
	if hasChildren {
		cmdPath = append(cmdPath, "readme")
	}
	cmdPath[len(cmdPath)-1] = cmdPath[len(cmdPath)-1] + ".md"

	filename := filepath.Join(dir, filepath.Join(cmdPath...))
	f, err := createFileAndDirs(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	prepend := fmt.Sprintf(`<!--
title: %s
-->
`, cmd.Name())

	if hasChildren {
		prepend = fmt.Sprintf(`<!--
title: %s
order: 0
-->
`, cmd.Name())
	}

	if _, err := io.WriteString(f, prepend); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f); err != nil {
		return err
	}
	return nil
}

func createFileAndDirs(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString("## " + name + "\n\n")
	buf.WriteString(cmd.Short + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}
