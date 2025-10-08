package main

import (
	"net/http"
)

const gameScript = `const ROWS = 6;
const COLS = 7;

const boardState = Array.from({ length: ROWS }, function () {
  return Array(COLS).fill(null);
});

let currentPlayer = "red";
let gridEl;
let statusEl;

function initElements() {
  gridEl = document.querySelector(".grid");
  statusEl = document.querySelector(".status");
}

function createBoard() {
  if (!gridEl) {
    return;
  }

  gridEl.innerHTML = "";
  for (let row = 0; row < ROWS; row += 1) {
    for (let col = 0; col < COLS; col += 1) {
      const cell = document.createElement("div");
      cell.className = "cell";
      cell.dataset.row = String(row);
      cell.dataset.col = String(col);
      cell.setAttribute(
        "aria-label",
        "Case ligne " + (row + 1) + ", colonne " + (col + 1)
      );
      gridEl.appendChild(cell);
    }
  }
}

function updateStatus() {
  if (!statusEl) {
    return;
  }

  const playerName = currentPlayer === "red" ? "rouge" : "jaune";
  statusEl.textContent = "Tour du joueur " + playerName;
  statusEl.dataset.player = currentPlayer;
}

function dropPiece(col) {
  if (boardState[0][col]) {
    return;
  }

  for (let row = ROWS - 1; row >= 0; row -= 1) {
    if (!boardState[row][col]) {
      const player = currentPlayer;
      boardState[row][col] = player;
      const selector =
        '.cell[data-row="' + row + '"][data-col="' + col + '"]';
      const targetCell = gridEl?.querySelector(selector);
      if (!targetCell) {
        return;
      }
      placePiece(targetCell, row, player);
      if (boardState[0][col]) {
        markColumnFull(col);
      }
      switchPlayer();
      updateStatus();
      break;
    }
  }
}

function markColumnFull(col) {
  const columnCells = gridEl?.querySelectorAll(
    '.cell[data-col="' + col + '"]'
  );
  columnCells?.forEach(function (cell) {
    cell.classList.add("is-full");
  });
}

function placePiece(cell, row, player) {
  const piece = document.createElement("div");
  piece.className = "piece piece--" + player + " is-dropping";
  piece.style.setProperty("--drop-distance", String(row + 1));
  cell.appendChild(piece);

  piece.addEventListener(
    "animationend",
    function () {
      piece.classList.remove("is-dropping");
    },
    { once: true }
  );
}

function switchPlayer() {
  currentPlayer = currentPlayer === "red" ? "yellow" : "red";
}

function handleGridClick(event) {
  const cell = event.target.closest(".cell");
  if (!cell) {
    return;
  }
  const col = Number.parseInt(cell.dataset.col ?? "", 10);
  if (Number.isNaN(col)) {
    return;
  }
  dropPiece(col);
}

function setupEventListeners() {
  gridEl?.addEventListener("click", handleGridClick);
}

function initGame() {
  initElements();
  createBoard();
  updateStatus();
  setupEventListeners();
}

document.addEventListener("DOMContentLoaded", initGame);
`

func registerGameRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/game.js", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = w.Write([]byte(gameScript))
	})
}
