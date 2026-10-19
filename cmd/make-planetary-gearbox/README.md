# make-planetary-gearbox — design notes

This document describes the physical tradeoffs of choosing different tooth
counts for the sun and planet gears, followed by concrete recommendations
(including a strong, compact 6:1 configuration) for
`make-planetary-gearbox`.

## The governing relationships

With the ring tooth count tied to the other two as `N_ring = N_sun + 2*N_planet`:

- **Center distance:** `a = module * (N_sun + N_planet) / 2`
- **Ratio with the ring fixed, sun input, carrier output:**

  ```
  reduction = 1 + N_ring / N_sun
            = 2 * (1 + N_planet / N_sun)
  ```

So **identical sun and planet gears always give exactly 4:1**. To change the
ratio you must make them different — there is no other degree of freedom once
the module and ring relation are fixed.

The assembly condition for three equally spaced planets is also affected:

```
3 divides (N_sun + N_planet)
```

which is why `-nt 13` alone fails but `-nt 13 -planetTeeth 14` works.

## Physical tradeoffs

### 1. Ratio, torque, and backdrivability

Larger planets relative to the sun (`N_planet/N_sun` increasing) increase the
reduction ratio and torque multiplication, but reduce backdrivability and
generally increase sliding losses. A smaller sun with larger planets gives the
highest ratio, but the sun becomes the weak link.

Identical gears lock you into 4:1. For example:

| Sun | Planet | Ring | Ratio | Character |
|---:|---:|---:|---:|---|
| 12 | 12 | 36 | 4.00 | compact, identical, but 12T is undercut-prone |
| 15 | 15 | 45 | 4.00 | same ratio, stronger, larger |
| 18 | 18 | 54 | 4.00 | strongest 4:1, largest |
| 13 | 14 | 41 | 4.15 | slightly higher ratio, co-prime counts |
| 13 | 11 | 35 | 3.69 | smaller/faster planets, weaker |
| 15 | 18 | 51 | 4.40 | higher ratio, larger ring |
| 18 | 15 | 48 | 3.67 | stronger sun, faster planets |

### 2. Undercutting and tooth strength

With a standard 20° involute and no profile shift, external gears below roughly
17 teeth are undercut; below about 12–14 they become noticeably weak and
pointed. The Lua node does not apply profile shift, so the **sun is usually the
limiting part** because it has the fewest teeth and carries the full input
torque. The planets are also at risk when they get small. Raising `-pa` to 25°
helps the undercut limit but increases radial bearing loads.

### 3. Contact ratio, smoothness, and noise

Higher tooth counts increase the contact ratio: more teeth share the load, the
mesh is smoother and quieter, and stress per tooth drops. Low-count
sun/planet pairs have lower contact ratio, more torque ripple, and more noise.
The internal planet/ring mesh usually has a high contact ratio because the ring
has many teeth.

### 4. Planet spin speed and bearing/wear

For a fixed sun speed, smaller planets spin faster relative to the carrier.
That means higher bearing/sliding speeds, more heat, and faster wear in the
print-in-place "bearing." Larger planets spin more slowly and are generally
more durable. Also, for a given input torque, a smaller sun raises the
tangential tooth force because the force is `T_sun / (n_planets * r_sun)`.

### 5. Ring diameter, print volume, and material

`N_ring = N_sun + 2*N_planet`, so larger planets grow the outer ring and the
whole print. For a fixed ring diameter, there is a direct trade:

- **More sun teeth, fewer planet teeth:** lower ratio, stronger sun, faster
  planets.
- **Fewer sun teeth, more planet teeth:** higher ratio, weaker sun, slower
  planets.

A fixed ring envelope also means you cannot freely choose all three counts;
the standard mesh relation ties them together.

### 6. Hunting tooth and wear distribution

Co-prime tooth counts distribute wear and reduce repeating noise. `13/14/41`
is good: `gcd(13,14)=1` and `gcd(14,41)=1`. `12/12/36` is poor: the same tooth
pairs tend to repeat. For a plastic toy this is secondary, but for durability
and quietness, co-prime counts are better.

### 7. Backlash and print tolerance

The absolute `-tol` clearance is a larger fraction of a small gear's tooth
size. Small-count sun/planets are more sensitive to print inaccuracy and may
need a slightly larger `-tol` or a larger module. Larger gears give more
forgiving geometry.

### 8. Internal-mesh interference

The ring profile is generated as the envelope of the mating planet, so it is
conjugate by construction. However, the planet tooth count still affects
tip/root clearances and the internal contact ratio. Very small planets approach
practical interference limits; larger planets generally make the internal mesh
more robust.

## Practical recommendations for this gearbox

- **Best all-around if the bed allows:** `-nt 15` → 15/15/45, exactly 4:1,
  stronger and less undercut than 12 or 13. `-nt 18` → 18/18/54 is even
  stronger but larger.
- **If you want to keep the 13-tooth sun:** use `-nt 13 -planetTeeth 14`
  (13/14/41, 4.15:1). It has co-prime counts, a stronger planet, slower planet
  spin, and a slightly higher ratio. Avoid `-planetTeeth 11` unless you
  specifically want a lower ratio and are willing to accept faster, weaker
  planets.
- **For a higher ratio:** increase the planet count relative to the sun, e.g.
  `-nt 15 -planetTeeth 18` (15/18/51, 4.4:1). Going much higher with a very
  small sun (e.g. 9/15/39, 5.33:1) gives a strong reduction but a weak,
  undercut sun.
- **For a lower ratio / more backdrivability:** use a larger sun and smaller
  planets, e.g. `-nt 18 -planetTeeth 15` (18/15/48, 3.67:1), but the planets
  spin faster.
- **Avoid below ~12 teeth** on the sun or planets at 20° pressure angle unless
  you accept undercut and reduced strength. If you must, consider `-pa 25`.

In short: identical gears buy simplicity and a fixed 4:1 ratio; different
counts let you trade ratio against sun strength, planet speed, ring size,
contact ratio, and wear. For a durable print-in-place unit, the sweet spot is
usually **15-tooth or larger sun and planets**, with counts chosen to satisfy
`3 divides (N_sun + N_planet)` and, ideally, co-prime.

## Strong, small 6:1 recommendations

For a 6:1 reduction with the ring fixed:

```
reduction = 1 + N_ring / N_sun = 6
=> N_ring = 5 * N_sun
```

Combining that with `N_ring = N_sun + 2*N_planet` gives:

```
N_planet = 2 * N_sun
N_ring   = 5 * N_sun
```

Consequences:

- Because `N_sun + N_planet = 3*N_sun`, the three-planet assembly condition
  `3 divides (N_sun + N_planet)` is **automatically satisfied** for every sun
  tooth count.
- The planet spin speed is favorable: the planets turn at one quarter of the
  sun speed in the opposite direction, independent of `N_sun`. That is slower
  than the 4:1 identical-gear case, where the planets turn at half the sun
  speed.
- The ratios are fixed by the tooth counts, so the two ways to make the unit
  "strong and small" are:
  - choose the **largest module** your printer handles well (stronger teeth),
    and
  - choose the **smallest `N_sun`** that keeps the sun out of serious undercut
    (smaller ring, since `N_ring = 5*N_sun`).
- Hunting-tooth behavior is inherently poor at exactly 6:1: every count is a
  multiple of `N_sun` (`gcd(N_sun, 2*N_sun, 5*N_sun) = N_sun`). This is fine
  for a printed toy; if wear/noise matter in a hard-use application, use a
  slightly different ratio with co-prime counts, or a two-stage gearbox.
- The ring reaction torque is large: at 6:1 the ring must react roughly seven
  times the sun input torque. Make sure the outer shell is mounted solidly and
  printed with enough wall thickness (`-outerThickness`).

Approximate outer diameter:

```
outer_diameter = module * N_ring
               + 2 * module
               + 2 * clearance
               + 2 * outer_thickness
```

### Recommended configurations

| Sun | Planet | Ring | Ratio | Module | Approx. outer Ø with `-tol 0.25 -outerThickness 8` | Character |
|---:|---:|---:|---:|---:|---:|---|
| 12 | 24 | 60 | 6.00 | 2.0 | 141 mm | smallest 6:1, but the 12T sun is undercut-prone |
| **15** | **30** | **75** | **6.00** | **1.5** | **132 mm** | **recommended strong/small balance** |
| 15 | 30 | 75 | 6.00 | 2.0 | 171 mm | stronger teeth, larger |
| 17 | 34 | 85 | 6.00 | 1.5 | 147 mm | no undercut on the sun, stronger |
| 17 | 34 | 85 | 6.00 | 1.25 | 125 mm | compact, but fine teeth (resin-friendly) |

### Recommended command (strong, small, 6:1)

```bash
go run ./cmd/make-planetary-gearbox \
  -nt 15 -planetTeeth 30 \
  -mod 1.5 -gl 20 \
  -tol 0.25 -outerThickness 8
```

This produces **15/30/75 teeth**, exactly **6:1**, with a ring outer diameter
of roughly **132 mm**. The 30-tooth planets are strong and slow-spinning; the
15-tooth sun is a reasonable compromise between compactness and undercut.

If FDM strength matters more than compactness, increase the module:

```bash
go run ./cmd/make-planetary-gearbox \
  -nt 15 -planetTeeth 30 \
  -mod 2 -gl 20 \
  -tol 0.25 -outerThickness 8
```

If you want to avoid sun undercut entirely, use 17 sun teeth:

```bash
go run ./cmd/make-planetary-gearbox \
  -nt 17 -planetTeeth 34 \
  -mod 1.5 -gl 20 \
  -tol 0.25 -outerThickness 8
```

That gives **17/34/85 teeth**, exactly **6:1**, with a roughly **147 mm** outer
diameter. For a compact resin print, `-mod 1.25` and `-outerThickness 6` bring
it down to roughly **121 mm**, at the cost of finer teeth.

### Notes for the 6:1 build

- `-ne 4` is the default; the number of herringbone elbows does not change the
  ratio.
- `-tol 0.25` is the target clearance. The measured minimum over the
  herringbone face width is about 0.21 mm for the 6:1 15/30/75 set (and about
  0.15 mm for the compact 12/12/36 set) because the radial clearance offset is
  less effective at the ring tooth tips. If you want a guaranteed 0.25 mm
  minimum, use `-tol 0.35`–`0.4`; `-tol 0.3` is a good compromise.
- The ring's `-outerThickness` is measured from the deepest notch to the outer
  smooth surface. For a 6:1 ring, keep it at least 6–8 mm for stiffness; 10 mm
  is safer for FDM.
- A 6:1 unit is not very backdrivable. It is good for holding torque and for
  driving an output from the sun; driving it backward from the carrier will be
  difficult.
