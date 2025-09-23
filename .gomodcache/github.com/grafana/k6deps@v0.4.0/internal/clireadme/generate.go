package clireadme

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func generateMarkdown(cmd *cobra.Command, w io.Writer, offset int) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buff := new(bytes.Buffer)

	printName(buff, cmd, offset)
	printShort(buff, cmd)
	printLong(buff, cmd, offset)
	printUseLine(buff, cmd)
	printExamples(buff, cmd, offset)
	printAdditionalHelpTopics(buff, cmd, offset)
	printFlags(buff, cmd, offset)
	printSeeAlso(buff, cmd, offset)

	if err := walkSubCommands(buff, cmd, offset); err != nil {
		return err
	}

	_, err := buff.WriteTo(w)
	return err
}

func heading(offset, level int, value string) string {
	return strings.Repeat("#", offset+level) + " " + value + "\n\n"
}

func getChildren(cmd *cobra.Command) []*cobra.Command {
	children := cmd.Commands()
	sort.Sort(byName(children))

	return children
}

func printName(buff *bytes.Buffer, cmd *cobra.Command, offset int) {
	buff.WriteString(heading(offset, 1, cmd.CommandPath()))
}

func printShort(buff *bytes.Buffer, cmd *cobra.Command) {
	if cmd.Runnable() {
		buff.WriteString(cmd.Short + "\n\n")
	} else {
		buff.WriteString("**" + cmd.Short + "**\n\n")
	}
}

func printLong(buff *bytes.Buffer, cmd *cobra.Command, offset int) {
	if len(cmd.Long) == 0 {
		return
	}

	if cmd.Runnable() {
		buff.WriteString(heading(offset, 2, "Synopsis"))
	}

	buff.WriteString(cmd.Long + "\n\n")
}

func printUseLine(buff *bytes.Buffer, cmd *cobra.Command) {
	if cmd.Runnable() {
		buff.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}
}

func printFlags(buff *bytes.Buffer, cmd *cobra.Command, offset int) {
	if !cmd.Runnable() {
		return
	}

	formatFlags := func(flags *pflag.FlagSet, title string) {
		flags.SetOutput(buff)
		if flags.HasAvailableFlags() {
			buff.WriteString(heading(offset, 2, title))
			buff.WriteString("```\n")
			flags.PrintDefaults()
			buff.WriteString("```\n\n")
		}
	}

	formatFlags(cmd.NonInheritedFlags(), "Flags")
	formatFlags(cmd.InheritedFlags(), "Inherited Flags")
}

func printExamples(buff *bytes.Buffer, cmd *cobra.Command, offset int) {
	if len(cmd.Example) == 0 {
		return
	}

	buff.WriteString(heading(offset, 2, "Examples"))
	buff.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
}

func printAdditionalHelpTopics(buff *bytes.Buffer, cmd *cobra.Command, offset int) {
	for _, child := range cmd.Commands() {
		if !child.IsAdditionalHelpTopicCommand() {
			continue
		}

		buff.WriteString(heading(offset, 2, child.Use))
		if len(child.Long) > 0 {
			buff.WriteString(child.Long + "\n\n")
		}
	}
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}

	for _, c := range cmd.Commands() {
		if c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand() {
			return true
		}
	}

	return false
}

func printSeeAlso(buf *bytes.Buffer, cmd *cobra.Command, offset int) {
	if !hasSeeAlso(cmd) {
		return
	}

	if cmd.HasParent() {
		buf.WriteString(heading(offset, 2, "SEE ALSO"))
		parent := cmd.Parent()
		pname := parent.CommandPath()
		link := strings.ReplaceAll(pname, " ", "-")
		buf.WriteString(fmt.Sprintf("* [%s](#%s)\t - %s\n", pname, link, parent.Short))
	}

	firstChild := true
	for _, child := range getChildren(cmd) {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}

		if firstChild {
			buf.WriteString(heading(offset, 2, "Commands"))
			firstChild = false
		}

		cname := cmd.CommandPath() + " " + child.Name()
		link := strings.ReplaceAll(cname, " ", "-")
		buf.WriteString(fmt.Sprintf("* [%s](#%s)\t - %s\n", cname, link, child.Short))
	}

	buf.WriteString("\n")
}

func walkSubCommands(buff *bytes.Buffer, cmd *cobra.Command, offset int) error {
	for _, child := range getChildren(cmd) {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}

		buff.WriteString("---\n")

		if err := generateMarkdown(child, buff, offset); err != nil {
			return err
		}
	}

	return nil
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
