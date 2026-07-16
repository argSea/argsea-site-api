package service

import "github.com/argSea/argsea-site-api/argHex/domain"

// seedCarvings is the shipped builtin carvings, one per spot. The first seven
// are the v1s, transcribed byte-for-byte from the design mock's svgCatalog
// (design/Admin.dc.html in argsea-site; the mock is design canon). The rest
// are the 2026-07-16 promote wave, lifted byte-verbatim from argsea-site
// main's built markup (index.astro, ShipsLog.tsx, 404.astro, TabBar.astro)
// and wrapped as standalone svgs with each mount's viewBox. Each seed is
// pre-bolted to its own spot: the current look on the site IS the builtin
// bolt, so "go back to the builtin" for a spot means re-bolting its seed.
// Animations never ride a carving: the records carry static geometry plus a
// data-carving-anchor tag where the page CSS keys a pulse on it (the lamp
// pattern); the two tab-bar seeds keep currentColor so the tab tint reaches
// the builtin, exactly as the tab-bar lighthouse does.
func seedCarvings() []domain.Carving {
	return []domain.Carving{
		{
			Name:     "The lighthouse",
			Svg:      `<svg width="24" height="28" viewBox="0 0 26 30" fill="none"><path d="M13 2 L17 9 L9 9 Z" fill="#f0d9a8"></path><rect x="10" y="9" width="6" height="14" fill="none" stroke="#93a0e8" stroke-width="1.4"></rect><path d="M10 13 h6 M10 17 h6" stroke="#93a0e8" stroke-width="1.4"></path><path d="M6 27 q7 -4 14 0" stroke="#5f6ec4" stroke-width="1.4" fill="none"></path></svg>`,
			BoltedTo: []string{domain.SpotLighthouseLogo},
		},
		{
			Name:     "The little boat",
			Svg:      `<svg width="30" height="24" viewBox="0 0 30 24" fill="none"><path d="M4 15 L26 15 L21 22 L9 22 Z" fill="#93a0e8"></path><path d="M15 15 V3" stroke="#5f6ec4" stroke-width="1.5"></path><path d="M15 3 L24 13 L15 13 Z" fill="#f0d9a8"></path></svg>`,
			BoltedTo: []string{domain.SpotBoat},
		},
		{
			Name:     "Message in a bottle",
			Svg:      `<svg width="32" height="20" viewBox="0 0 40 24" fill="none"><rect x="6" y="7" width="28" height="11" rx="5.5" fill="rgba(147,160,232,.22)" stroke="#93a0e8" stroke-width="1.3"></rect><rect x="33" y="9.5" width="5" height="6" rx="1.2" fill="#f0d9a8"></rect><path d="M12 10 h14 M12 12.5 h11 M12 15 h13" stroke="#f0d9a8" stroke-width="1" stroke-linecap="round" opacity=".85"></path></svg>`,
			BoltedTo: []string{domain.SpotBottle},
		},
		{
			Name:     "Tower on the horizon",
			Svg:      `<svg width="26" height="34" viewBox="0 0 26 34" fill="none"><path d="M13 3 L17 10 L9 10 Z" fill="rgba(150,160,220,.4)"></path><rect x="10" y="10" width="6" height="15" fill="none" stroke="rgba(150,160,220,.45)" stroke-width="1.3"></rect><path d="M10 14 h6 M10 19 h6" stroke="rgba(150,160,220,.34)" stroke-width="1.1"></path><path d="M5 30 q8 -4 16 0" stroke="rgba(150,160,220,.36)" stroke-width="1.3" fill="none"></path></svg>`,
			BoltedTo: []string{domain.SpotTowerStub},
		},
		{
			Name:     "Paw print",
			Svg:      `<svg width="13" height="12" viewBox="0 0 15 14" fill="#93a0e8"><ellipse cx="7.5" cy="9.5" rx="3.4" ry="2.9"></ellipse><ellipse cx="2.6" cy="5.4" rx="1.5" ry="1.9"></ellipse><ellipse cx="6.2" cy="3.4" rx="1.5" ry="1.9"></ellipse><ellipse cx="9.8" cy="3.6" rx="1.5" ry="1.9"></ellipse><ellipse cx="12.6" cy="6" rx="1.4" ry="1.8"></ellipse></svg>`,
			BoltedTo: []string{domain.SpotPaw},
		},
		{
			Name:     "The wave line",
			Svg:      `<svg xmlns="http://www.w3.org/2000/svg" width="53" height="18"><path d="M0 9 Q 13.25 0, 26.5 9 T 53 9" stroke="rgba(147,160,232,0.5)" stroke-width="1.5" fill="none"/></svg>`,
			BoltedTo: []string{domain.SpotWaveLine},
		},
		{
			Name:     "The boat wake",
			Svg:      `<svg xmlns="http://www.w3.org/2000/svg" width="53" height="18"><path d="M0 9 Q 13.25 0, 26.5 9 T 53 9" stroke="rgba(240,217,168,0.5)" stroke-width="1.5" fill="none"/></svg>`,
			BoltedTo: []string{domain.SpotBoatWake},
		},
		{
			Name:     "The morse seal",
			Svg:      `<svg width="34" height="38" viewBox="0 0 18 20" fill="none" stroke="#93a0e8" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"><path data-carving-anchor="lamp" d="M9 2 L12 6 L6 6 Z" fill="#93a0e8" stroke="none"></path><path d="M7 6 L6 15 M11 6 L12 15 M7 9.5 h4"></path><path d="M4 17 q5 -3 10 0"></path></svg>`,
			BoltedTo: []string{domain.SpotMorseSeal},
		},
		{
			Name:     "The panel rose",
			Svg:      `<svg width="36" height="36" viewBox="0 0 36 36" fill="none"><circle cx="18" cy="18" r="8.5" stroke="#93a0e8" stroke-width=".9" opacity=".5"></circle><circle cx="18" cy="18" r="12.5" stroke="#93a0e8" stroke-width=".6" opacity=".25" stroke-dasharray="1.5 3"></circle><path d="M18 2 L20 16 L18 18 L16 16 Z" fill="#93a0e8" opacity=".8"></path><path d="M18 34 L20 20 L18 18 L16 20 Z" fill="#93a0e8" opacity=".4"></path><path d="M2 18 L16 16 L18 18 L16 20 Z" fill="#93a0e8" opacity=".4"></path><path d="M34 18 L20 16 L18 18 L20 20 Z" fill="#93a0e8" opacity=".4"></path><circle cx="18" cy="18" r="1.2" fill="#93a0e8"></circle></svg>`,
			BoltedTo: []string{domain.SpotPanelRose},
		},
		{
			Name:     "The fleet wake",
			Svg:      `<svg viewBox="0 0 800 60" preserveAspectRatio="none"><path d="M-8 44 C 110 12, 230 56, 380 30 S 640 4, 808 26" fill="none" stroke="rgba(147,160,232,.4)" stroke-width="1.4" stroke-linecap="round" stroke-dasharray="0.1 8"></path></svg>`,
			BoltedTo: []string{domain.SpotFleetWake},
		},
		{
			Name:     "The sea serpent",
			Svg:      `<svg width="72" height="26" viewBox="0 0 72 26" fill="none"><path d="M2 16 q7 -12 14 0 t14 0 t14 0 t14 0" stroke="#6a76c8" stroke-width="1.6" fill="none" stroke-linecap="round" /><circle cx="66" cy="9" r="2.4" fill="#6a76c8" /><path d="M69 8 l4 -3 M69 10 l4 2" stroke="#6a76c8" stroke-width="1.2" stroke-linecap="round" /></svg>`,
			BoltedTo: []string{domain.SpotSeaSerpent},
		},
		{
			Name:     "The marooned palm",
			Svg:      `<svg width="30" height="30" viewBox="0 0 30 30" fill="none"><path d="M14 28 q-2 -12 1 -20" stroke="#8a7142" stroke-width="2" fill="none" stroke-linecap="round" /><path d="M15 8 q-8 -3 -13 1 M15 8 q8 -3 13 1 M15 8 q-5 -6 -12 -6 M15 8 q5 -6 12 -6" stroke="#5f8a5f" stroke-width="1.8" fill="none" stroke-linecap="round" /></svg>`,
			BoltedTo: []string{domain.SpotMaroonedPalm},
		},
		{
			Name:     "The signal flare",
			Svg:      `<svg width="13" height="15" viewBox="0 0 13 15" fill="none"><path d="M6.5 14 V6" stroke="#8a6d3b" stroke-width="1.3" stroke-linecap="round" /><path d="M6.5 1 L8 4.5 L6.5 3.5 L5 4.5 Z" fill="#d64535" /><circle cx="6.5" cy="2" r="1.4" fill="#ff6a52" /></svg>`,
			BoltedTo: []string{domain.SpotSignalFlare},
		},
		{
			Name:     "The port anchor",
			Svg:      `<svg width="26" height="30" viewBox="0 0 26 30" fill="none"><circle cx="13" cy="5" r="3" stroke="#93a0e8" stroke-width="1.6" /><path d="M13 8 V26" stroke="#93a0e8" stroke-width="1.6" /><path d="M6 15 H20" stroke="#93a0e8" stroke-width="1.6" /><path d="M5 22 q8 7 16 0" stroke="#93a0e8" stroke-width="1.6" fill="none" stroke-linecap="round" /></svg>`,
			BoltedTo: []string{domain.SpotPortAnchor},
		},
		{
			Name:     "The chart rose",
			Svg:      `<svg width="104" height="104" viewBox="0 0 104 104" fill="none"><circle cx="52" cy="52" r="34" stroke="rgba(147,160,232,.35)" stroke-width="1" /><g opacity=".9"><path d="M52 8 L58 52 L52 46 L46 52 Z" fill="#f0d9a8" /><path d="M52 96 L46 52 L52 58 L58 52 Z" fill="#5f6ec4" /><path d="M8 52 L52 46 L46 52 L52 58 Z" fill="#5f6ec4" /><path d="M96 52 L52 58 L58 52 L52 46 Z" fill="#5f6ec4" /></g><g opacity=".55"><path d="M52 52 L74 30 L58 50 Z" fill="#93a0e8" /><path d="M52 52 L74 74 L54 58 Z" fill="#6a76c8" /><path d="M52 52 L30 74 L46 54 Z" fill="#93a0e8" /><path d="M52 52 L30 30 L50 46 Z" fill="#6a76c8" /></g><circle cx="52" cy="52" r="3.4" fill="#f0d9a8" /></svg>`,
			BoltedTo: []string{domain.SpotChartRose},
		},
		{
			Name:     "The compass-rose star",
			Svg:      `<svg width="32" height="32" viewBox="0 0 30 30" fill="none"><path d="M15 0 L17 13 L15 11 L13 13 Z M15 30 L13 17 L15 19 L17 17 Z M0 15 L13 13 L11 15 L13 17 Z M30 15 L17 17 L19 15 L17 13 Z" fill="#ff6a52" /><circle cx="15" cy="15" r="3" fill="#fff" /></svg>`,
			BoltedTo: []string{domain.SpotCompassRoseStar},
		},
		{
			Name:     "The sail tent",
			Svg:      `<svg width="24" height="15" viewBox="0 0 22 14"><path d="M2 12 Q11 -6 20 12" fill="none" stroke="rgba(255,106,82,.75)" stroke-width="1.4" /><path d="M11 1 V12 M2 12 l9 -3 M20 12 l-9 -3" stroke="rgba(255,106,82,.5)" stroke-width="1" /></svg>`,
			BoltedTo: []string{domain.SpotSailTent},
		},
		{
			Name:     "The moored lamp",
			Svg:      `<svg width="30" height="34" viewBox="0 0 26 30" fill="none"><path d="M13 2 L17 9 L9 9 Z" fill="#f0d9a8" data-carving-anchor="lamp" /><rect x="10" y="9" width="6" height="14" fill="none" stroke="#93a0e8" stroke-width="1.4" /><path d="M10 13 h6 M10 17 h6" stroke="#93a0e8" stroke-width="1.4" /></svg>`,
			BoltedTo: []string{domain.SpotMooredLamp},
		},
		{
			Name:     "The adrift boat",
			Svg:      `<svg width="34" height="28" viewBox="0 0 30 24" fill="none"><path d="M4 15 L26 15 L21 22 L9 22 Z" fill="#93a0e8" /><path d="M15 15 V3" stroke="#5f6ec4" stroke-width="1.5" /><path d="M15 3 L24 13 L15 13 Z" fill="#f0d9a8" /></svg>`,
			BoltedTo: []string{domain.SpotAdriftBoat},
		},
		{
			Name:     "The adrift wake",
			Svg:      `<svg width="52" height="16" viewBox="0 0 60 16" fill="none"><path d="M2 8 q7 -6 14 0 t14 0 t14 0 t14 0" stroke="rgba(240,217,168,.5)" stroke-width="1.4" fill="none" stroke-dasharray="1 5" stroke-linecap="round" /></svg>`,
			BoltedTo: []string{domain.SpotAdriftWake},
		},
		{
			Name:     "The gull",
			Svg:      `<svg width="26" height="10" viewBox="0 0 26 10" fill="none"><path d="M1 8 Q6.5 1 13 6 Q19.5 1 25 8" stroke="#8f9be0" stroke-width="1.4" fill="none" stroke-linecap="round" /></svg>`,
			BoltedTo: []string{domain.SpotGull},
		},
		{
			Name:     "The route line",
			Svg:      `<svg viewBox="0 0 300 70" fill="none" preserveAspectRatio="none"><path d="M0 6 Q90 -8 170 34 T292 58" stroke="rgba(240,217,168,.5)" stroke-width="2" stroke-dasharray="2 7" stroke-linecap="round" fill="none" /></svg>`,
			BoltedTo: []string{domain.SpotRouteLine},
		},
		{
			Name:     "The buoy",
			Svg:      `<svg width="30" height="46" viewBox="0 0 30 46" fill="none"><ellipse cx="15" cy="8" rx="4" ry="4" fill="#f0d9a8" /><path d="M15 12 V20" stroke="#5f6ec4" stroke-width="1.6" /><path d="M6 22 L24 22 L21 40 L9 40 Z" fill="#c05a4a" /><path d="M7.4 30 L22.6 30" stroke="#f0d9a8" stroke-width="4" /><path d="M6 22 L24 22" stroke="#e8ebfa" stroke-width="1.4" opacity=".6" /></svg>`,
			BoltedTo: []string{domain.SpotBuoy},
		},
		{
			Name:     "The compass",
			Svg:      `<svg width="20" height="20" viewBox="0 0 22 22" fill="none"><circle cx="11" cy="11" r="8.5" stroke="currentColor" stroke-width="1.4"></circle><path d="M11 3 L12.4 11 L11 19 L9.6 11 Z" fill="currentColor"></path><path d="M3 11 L11 9.6 L19 11 L11 12.4 Z" fill="currentColor" opacity="0.55"></path></svg>`,
			BoltedTo: []string{domain.SpotCompass},
		},
		{
			Name:     "The notes letter",
			Svg:      `<svg width="19" height="22" viewBox="0 0 20 22" fill="none"><rect x="4" y="3" width="12" height="16" rx="1.5" stroke="currentColor" stroke-width="1.5"></rect><path d="M7 7 h6 M7 10 h6 M7 13 h4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"></path></svg>`,
			BoltedTo: []string{domain.SpotNotesLetter},
		},
	}
}
