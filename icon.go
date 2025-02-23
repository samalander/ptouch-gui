package main

import "fyne.io/fyne/v2"

var resourceIconSvg = &fyne.StaticResource{
    StaticName: "icon.svg",
    StaticContent: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
    <rect x="12" y="20" width="40" height="30" fill="#505050" rx="4"/>
    <rect x="10" y="28" width="44" height="4" fill="#404040"/>
    <circle cx="48" cy="38" r="3" fill="#00ff00"/>
    <rect x="16" y="38" width="8" height="2" fill="#808080"/>
    <path d="M10 30 L54 30 L54 28 L10 28 L10 30" fill="#ffffff"/>
    <path d="M15 20 L49 20 L49 12 L15 12 Z" fill="#606060"/>
</svg>`),
}
