package main

import (
   "encoding/json"
   "fmt"
   "io"
   "log"
   "net/http"
   "sort"
   "strings"

   "github.com/gdamore/tcell/v2"
   "github.com/rivo/tview"
)

const apiURL = "https://pokeapi.co/api/v2"

// PokemonListEntry represents a basic Pokémon entry from the API.
type PokemonListEntry struct {
   Name string `json:"name"`
   URL  string `json:"url"`
}

// PokemonListResponse is the JSON response for listing Pokémon.
type PokemonListResponse struct {
   Results []PokemonListEntry `json:"results"`
}

// PokemonType represents a Pokémon type.
type PokemonType struct {
   Type struct {
       Name string `json:"name"`
   } `json:"type"`
}

// PokemonAbility represents a Pokémon ability.
type PokemonAbility struct {
   Ability struct {
       Name string `json:"name"`
   } `json:"ability"`
}

// PokemonStat represents a base stat of a Pokémon.
type PokemonStat struct {
   Stat struct {
       Name string `json:"name"`
   } `json:"stat"`
   BaseStat int `json:"base_stat"`
}

// PokemonDetails represents detailed information about a Pokémon.
type PokemonDetails struct {
   Name           string           `json:"name"`
   ID             int              `json:"id"`
   Height         int              `json:"height"`
   Weight         int              `json:"weight"`
   BaseExperience int              `json:"base_experience"`
   Types          []PokemonType    `json:"types"`
   Abilities      []PokemonAbility `json:"abilities"`
   Stats          []PokemonStat    `json:"stats"`
}

// fetchPokemonList fetches all Pokémon and groups them by first letter.
func fetchPokemonList() ([]*tview.TreeNode, error) {
   resp, err := http.Get(fmt.Sprintf("%s/pokemon?limit=10000", apiURL))
   if err != nil {
       return nil, err
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
       return nil, fmt.Errorf("status code %d", resp.StatusCode)
   }
   body, err := io.ReadAll(resp.Body)
   if err != nil {
       return nil, err
   }
   var listResp PokemonListResponse
   if err := json.Unmarshal(body, &listResp); err != nil {
       return nil, err
   }

   groups := make(map[string][]PokemonListEntry)
   for _, p := range listResp.Results {
       letter := strings.ToUpper(string(p.Name[0]))
       groups[letter] = append(groups[letter], p)
   }

   letters := make([]string, 0, len(groups))
   for l := range groups {
       letters = append(letters, l)
   }
   sort.Strings(letters)

   var nodes []*tview.TreeNode
   for _, l := range letters {
       parent := tview.NewTreeNode(l).
           SetColor(tcell.ColorYellow).
           SetExpanded(true)
       entries := groups[l]
       sort.Slice(entries, func(i, j int) bool {
           return entries[i].Name < entries[j].Name
       })
       for _, entry := range entries {
           node := tview.NewTreeNode(entry.Name).
               SetReference(entry).
               SetSelectable(true)
           parent.AddChild(node)
       }
       nodes = append(nodes, parent)
   }
   return nodes, nil
}

// fetchPokemonDetails fetches detailed info for a given Pokémon URL.
func fetchPokemonDetails(url string) (PokemonDetails, error) {
   resp, err := http.Get(url)
   if err != nil {
       return PokemonDetails{}, err
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
       return PokemonDetails{}, fmt.Errorf("status code %d", resp.StatusCode)
   }
   body, err := io.ReadAll(resp.Body)
   if err != nil {
       return PokemonDetails{}, err
   }
   var details PokemonDetails
   if err := json.Unmarshal(body, &details); err != nil {
       return PokemonDetails{}, err
   }
   return details, nil
}

func main() {
   app := tview.NewApplication()

   // Tree view for Pokémon list.
   tree := tview.NewTreeView()
   tree.SetBorder(true)
   tree.SetTitle("Pokémons")
   root := tview.NewTreeNode("Pokémons").SetColor(tcell.ColorGreen)
   tree.SetRoot(root)
   tree.SetCurrentNode(root)

   // Text view for details.
   details := tview.NewTextView()
   details.SetBorder(true)
   details.SetTitle("Details")
   details.SetDynamicColors(true)
   details.SetWrap(true)
   details.SetScrollable(true)

   // Load list.
   nodes, err := fetchPokemonList()
   if err != nil {
       log.Fatalf("Failed to load list: %v", err)
   }
   for _, n := range nodes {
       root.AddChild(n)
   }
   root.SetExpanded(true)

   // Selection handler.
   tree.SetSelectedFunc(func(node *tview.TreeNode) {
       ref := node.GetReference()
       if ref == nil {
           node.SetExpanded(!node.IsExpanded())
           return
       }
       entry := ref.(PokemonListEntry)
       details.Clear()
       fmt.Fprintf(details, "Loading %s...\n", entry.Name)
       go func() {
           detail, err := fetchPokemonDetails(entry.URL)
           app.QueueUpdateDraw(func() {
               details.Clear()
               if err != nil {
                   fmt.Fprintf(details, "[red]Error:[-] %v", err)
                   return
               }
               // Prepare types and abilities
               var types []string
               for _, t := range detail.Types {
                   types = append(types, t.Type.Name)
               }
               var abilities []string
               for _, a := range detail.Abilities {
                   abilities = append(abilities, a.Ability.Name)
               }
               fmt.Fprintf(details,
                   "Name: %s\nID: %d\nHeight: %d\nWeight: %d\nBase XP: %d\nTypes: %s\nAbilities: %s\nStats:\n",
                   detail.Name, detail.ID, detail.Height, detail.Weight, detail.BaseExperience,
                   strings.Join(types, ", "),
                   strings.Join(abilities, ", "),
               )
               for _, s := range detail.Stats {
                   fmt.Fprintf(details, "  %s: %d\n", s.Stat.Name, s.BaseStat)
               }
           })
       }()
   })

   // Global key binding: quit on 'q' or 'Q'.
   app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
       if event.Key() == tcell.KeyRune {
           switch event.Rune() {
           case 'q', 'Q':
               app.Stop()
               return nil
           }
       }
       return event
   })

   // Layout: tree and details side by side.
   layout := tview.NewFlex().
       AddItem(tree, 0, 1, true).
       AddItem(details, 0, 2, false)

   if err := app.SetRoot(layout, true).SetFocus(tree).Run(); err != nil {
       log.Fatalf("Application error: %v", err)
   }
}