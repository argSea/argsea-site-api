package service

import "github.com/argSea/argsea-site-api/argHex/domain"

// seedCarvings is the seven shipped v1 carvings, one per spot, transcribed
// byte-for-byte from the design mock's svgCatalog (design/Admin.dc.html in
// argsea-site; the mock is design canon). Each seed is pre-bolted to its own
// spot: the current look on the site IS the v1 bolt, so "go back to v1" for a
// spot means re-bolting its seed.
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
	}
}
