# PixelForge

Welcome to **PixelForge**, the home of all game source code for **PixelArcade**! Each folder in this directory 
represents an individual game that is part of our arcade. Developers can contribute new games, update existing 
ones, and explore how each game is built.

## Folder Structure
Each game should have its own dedicated folder inside **PixelForge**. The folder name should match the game's
title and contain all necessary source code, assets, and dependencies specific to that game.

Example:
```
PixelForge/
├── retro_racer/
│   ├── assets/
│   ├── src/
│   ├── build.sh
│   └── README.md
├── space_blaster/
│   ├── assets/
│   ├── src/
│   ├── build.sh
│   └── README.md
```

## Game-Specific README
Each game folder **must** include its own `README.md` file with the following details:
- **Game Description** – A brief overview of the game.
- **How to Build** – Steps to compile or package the game.
- **How to Play** – Controls and gameplay mechanics.
- **Dependencies** – Any required libraries or tools.

## Contributing
Want to add a game to **PixelForge**? Follow these steps:
1. Create a new folder inside `PixelForge/` with your game's name.
2. Add all source code and assets inside that folder.
3. Write a `README.md` inside your game's folder with build instructions.
4. Submit a pull request or push your changes following the project's [contribution guidelines](../docs/CONTRIBUTING.md).

## Notes
- Each game is **independent** and should not rely on other games.
- Avoid conflicts by using unique asset and code namespaces.
- Ensure your game works in a browser environment before submitting.

Happy coding, and welcome to **PixelForge**! 🚀🎮

