package service

import "github.com/argSea/argsea-site-api/argHex/domain"

// The seeded v1 designs are the two shipped harbor cats, translated shape for
// shape from the site's HarborCat island (design/HarborCat.dc.html and the
// approved lying hybrid). Every d string and coordinate is byte-identical to
// the shipped SVG — these must render exactly the cat that sails today, or
// "go back to v1" goes back to the wrong cat. Circles become ellipses with
// rx = ry: the Shape vocabulary has no circle.

// seedPerchedV1 is the cat perched on a corner, tail draped over the edge.
func seedPerchedV1() domain.CatDesign {
	return domain.CatDesign{
		Pose:    domain.PosePerched,
		Label:   "v1",
		ViewBox: "0 0 64 74",
		Shapes: []domain.Shape{
			{Id: "tail", Type: "path", D: "M45 55 C57 52 61 62 56 70 C54.5 72.5 51 72.5 50 70 C52.5 64.5 50 60 43 60 Z", Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.4, Linejoin: "round", Role: "tail", Origin: []float64{45, 56}},
			{Id: "paw-left", Type: "ellipse", Cx: 26, Cy: 52.5, Rx: 4.4, Ry: 3.2, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.3},
			{Id: "paw-right", Type: "ellipse", Cx: 37.5, Cy: 52.5, Rx: 4.4, Ry: 3.2, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.3},
			{Id: "body", Type: "path", D: "M12.95 51.8 C9.25 40.7 13.9 29.6 22.2 25.9 L21.3 15.7 L27.4 21.8 L33.7 21.8 L39.8 15.7 L38.85 25.9 C47.2 29.6 51.8 40.7 48.1 51.8 Z", Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.6, Linejoin: "round", Role: "body"},
			{Id: "ear-left", Type: "path", D: "M22.7 18.2 L25.6 21.6 L22.2 21.6 Z", Fill: "#f0d9a8", Opacity: 0.5},
			{Id: "ear-right", Type: "path", D: "M38.3 18.2 L38.8 21.6 L35.4 21.6 Z", Fill: "#f0d9a8", Opacity: 0.5},
			{Id: "eye-left", Type: "ellipse", Cx: 25.9, Cy: 30.8, Rx: 1.9, Ry: 1.9, Fill: "#f0d9a8", Role: "eyes", Origin: []float64{30, 31}},
			{Id: "eye-right", Type: "ellipse", Cx: 35.2, Cy: 30.8, Rx: 1.9, Ry: 1.9, Fill: "#f0d9a8", Role: "eyes", Origin: []float64{30, 31}},
			{Id: "nose", Type: "path", D: "M29.4 35 L32.2 35 L30.8 36.6 Z", Fill: "#f0d9a8"},
			{Id: "mouth", Type: "path", D: "M30.8 36.6 v1.4 M30.8 38 q-2 1.4 -3.6 .4 M30.8 38 q2 1.4 3.6 .4", Fill: "none", Stroke: "#5f6ec4", StrokeWidth: 1, Linecap: "round"},
			{Id: "whiskers", Type: "path", D: "M22 33 l-7 -1.4 M22 35.4 l-7 1 M39.5 33 l7 -1.4 M39.5 35.4 l7 1", Fill: "none", Stroke: "#5f6ec4", StrokeWidth: 0.9, Linecap: "round", Opacity: 0.7},
		},
	}
}

// seedLyingV1 is the cat draped along a horizontal element — the approved
// hybrid loaf, hind paw and chest stripe included.
func seedLyingV1() domain.CatDesign {
	return domain.CatDesign{
		Pose:    domain.PoseLying,
		Label:   "v1",
		ViewBox: "0 0 100 48",
		Shapes: []domain.Shape{
			{Id: "tail", Type: "path", D: "M72 38 C85 33 91 41 86 46.5 C84 49 80.2 48.4 80.8 45.2 C83 41.4 78.6 39.6 72.6 42.4 Z", Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.4, Linejoin: "round", Role: "tail", Origin: []float64{73, 40}},
			{Id: "paw-front-left", Type: "ellipse", Cx: 10.5, Cy: 42.6, Rx: 4.6, Ry: 2.7, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.3},
			{Id: "paw-front-right", Type: "ellipse", Cx: 18.5, Cy: 42.9, Rx: 4.4, Ry: 2.6, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.3},
			{Id: "body", Type: "path", D: "M13.5 44 C8.5 39 9 30 13 25 L12.8 22 L12.4 10 L18.6 17 L24.5 17 L30 10 L30.5 22 C35.5 25 38.5 27 44.5 28.5 C56 25.5 69 26.5 77.5 32.5 C83.5 36.8 83 42 76 44 Z", Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.6, Linejoin: "round", Role: "body"},
			{Id: "paw-hind", Type: "ellipse", Cx: 68, Cy: 43.2, Rx: 5, Ry: 2.4, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.2},
			{Id: "ear-left", Type: "path", D: "M13.7 16 L13.5 11.5 L17 15 Z", Fill: "#f0d9a8", Opacity: 0.5},
			{Id: "ear-right", Type: "path", D: "M29 16 L29.4 11.5 L26 15 Z", Fill: "#f0d9a8", Opacity: 0.5},
			{Id: "eye-left", Type: "ellipse", Cx: 17.8, Cy: 26.8, Rx: 1.9, Ry: 1.9, Fill: "#f0d9a8", Role: "eyes", Origin: []float64{22, 27}},
			{Id: "eye-right", Type: "ellipse", Cx: 26, Cy: 26.8, Rx: 1.9, Ry: 1.9, Fill: "#f0d9a8", Role: "eyes", Origin: []float64{22, 27}},
			{Id: "nose", Type: "path", D: "M20.6 30.4 L23.4 30.4 L22 32 Z", Fill: "#f0d9a8"},
			{Id: "mouth", Type: "path", D: "M22 32 v1.3 M22 33.3 q-2 1.4 -3.6 .4 M22 33.3 q2 1.4 3.6 .4", Fill: "none", Stroke: "#5f6ec4", StrokeWidth: 1, Linecap: "round"},
			{Id: "whiskers", Type: "path", D: "M12.6 29 l-7 -1.3 M12.6 31.2 l-7 .9 M30.5 29 l7 -1.3 M30.5 31.2 l7 .9", Fill: "none", Stroke: "#5f6ec4", StrokeWidth: 0.9, Linecap: "round", Opacity: 0.7},
			{Id: "chest-stripe", Type: "path", D: "M15.8 36 q3 1.6 6.4 .6", Fill: "none", Stroke: "#5f6ec4", StrokeWidth: 0.9, Linecap: "round", Opacity: 0.45},
		},
	}
}
