#!/usr/bin/env python3

import asyncio
import requests
from textual.app import App
from textual.widgets import Header, Footer, TreeControl, ScrollView, Static

API_URL = "https://pokeapi.co/api/v2"

class PokeApp(App):
    """A Textual TUI application to browse Pokémon using the PokeAPI."""

    TITLE = "PokeAPI TUI"
    BINDINGS = [("q", "quit", "Quit")]

    async def on_mount(self) -> None:
        """Mount the UI layout and load Pokémon list."""
        await self.view.dock(Header(), edge="top")
        await self.view.dock(Footer(), edge="bottom")
        self.tree = TreeControl("Pokémons", {})
        self.details = ScrollView(Static("Select a Pokémon from the list"))
        await self.view.dock(self.tree, edge="left", size=40)
        await self.view.dock(self.details, edge="right")
        await self.load_pokemon_list()

    async def load_pokemon_list(self) -> None:
        """Fetch list of all Pokémon and populate the tree view."""
        try:
            response = await asyncio.to_thread(requests.get, f"{API_URL}/pokemon?limit=10000")
            response.raise_for_status()
            data = response.json()
        except Exception as e:
            self.details.update(f"Failed to load Pokémon list: {e}")
            return

        pokemons = data.get("results", [])
        # Group by starting letter for easier navigation
        grouped = {}
        for p in pokemons:
            letter = p["name"][0].upper()
            grouped.setdefault(letter, []).append(p)

        for letter in sorted(grouped):
            nodes = sorted(grouped[letter], key=lambda x: x["name"])
            letter_node = await self.tree.add(self.tree.root, letter, expand=True)
            for p in nodes:
                await self.tree.add(letter_node, p["name"], data=p)
        await self.tree.root.expand()

    async def on_tree_node_selected(self, message) -> None:
        """Handle selection events on the tree view."""
        node = message.node
        if not node.data:
            return
        pokemon = node.data
        try:
            resp = await asyncio.to_thread(requests.get, pokemon["url"])
            resp.raise_for_status()
            details = resp.json()
        except Exception as e:
            self.details.update(f"Failed to load details for {pokemon['name']}: {e}")
            return

        lines = [
            f"Name: {details['name']}",
            f"ID: {details['id']}",
            f"Height: {details['height']}",
            f"Weight: {details['weight']}",
            f"Base experience: {details['base_experience']}",
            "Types: " + ", ".join(t["type"]["name"] for t in details["types"]),
            "Abilities: " + ", ".join(a["ability"]["name"] for a in details["abilities"]),
            "Stats:",
        ]
        for stat in details.get("stats", []):
            lines.append(f"  {stat['stat']['name']}: {stat['base_stat']}")

        self.details.update("\n".join(lines))

if __name__ == "__main__":
    PokeApp.run()