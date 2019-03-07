package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
	"text/template"
)

var (
	flags      *flag.FlagSet
	listSize   int
	listAll    bool
	listRemote bool
	hideHelp   bool
)

var rootCmd = &cobra.Command{
	Use:   "git checkout-branch",
	Short: "Checkout git branches more efficiently.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Help()
			os.Exit(1)
		}

		var branches []*Branch
		switch {
		case listAll:
			branches = allBranches()
		case listRemote:
			branches = remoteBranches()
		default:
			branches = localBranches()
		}

		if len(branches) == 0 {
			return
		}

		branch := selectBranch(branches, listSize, hideHelp)
		if branch != nil {
			checkoutBranch(branch)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&listRemote, "remotes", "r", false, "")
	rootCmd.Flags().BoolVarP(&listAll, "all", "a", false, "")
	rootCmd.Flags().IntVarP(&listSize, "number", "n", 10, "")
	rootCmd.Flags().BoolVarP(&hideHelp, "hide-help", "", false, "")

	rootCmd.SetUsageFunc(func(*cobra.Command) error {
		usage := `Usage:
  git checkout-branch [flags]

Flags:
  -a, --all          List both remote-tracking branches and local branches
  -r, --remotes      List the remote-tracking branches
  -n, --number       Set the number of branches displayed in the list (default 10)
      --hide-help    Hide the help information`
		fmt.Println(usage)
		os.Exit(1)
		return nil
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getName(name string) string {
	out := bytes.Buffer{}
	Input := promptui.CurrentCursor.Get()
	if Input != "" && strings.Contains(name, Input) {
		tpl, err := template.New("").Funcs(promptui.FuncMap).Parse("{{ . | green }}")
		if err != nil {
			fmt.Println(err.Error())
		}
		err1 := tpl.Execute(&out, Input)
		if err1 != nil {
			fmt.Println(err1.Error())
		}
		return strings.Replace(name, Input, out.String(), -1)
	} else {
		return name
	}
}

func selectBranch(branches []*Branch, size int, hideHelp bool) *Branch {
	iconSelect := promptui.Styler(promptui.FGGreen)("*")
	p := promptui.FuncMap
	p["getName"] = getName
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   iconSelect + " {{ .Index | green }} {{ .Name | green }}",
		Inactive: "  {{ .Index }} {{ getName .Name }}",
		FuncMap: p,
	}
	searcher := func(input string, index int) bool {
		b := branches[index]
		name := strings.Replace(strings.ToLower(b.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input) || input == b.Index
	}

	label := strconv.Itoa(len(branches)) + " Branches:"

	currentBranch := currentBranch()
	position := 0
	for index, item := range branches {
		if item.Name == currentBranch.Name {
			position = index
		}
	}

	prompt := promptui.Select{
		Label:        label,
		Items:        branches,
		Templates:    templates,
		Size:         size,
		Searcher:     searcher,
		HideHelp:     hideHelp,
		HideSelected: true,
		StartInSearchMode: true,
	}

	i, _, err := prompt.RunCursorAt(position, position - size / 2)
	if err != nil {
		return nil
	}
	return branches[i]
}
