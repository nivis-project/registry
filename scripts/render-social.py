#!/usr/bin/env python3
import re
import subprocess
import sys
import tempfile

from coloraide import Color

OKLCH = re.compile(r"oklch\([^)]*\)")


def to_srgb(match):
    return Color(match.group(0)).convert("srgb").fit("srgb").to_string(hex=True)


def main():
    if len(sys.argv) != 4:
        sys.exit("usage: render-social.py <social.svg> <social.png> <fonts-dir>")
    svg_path, png_path, fonts = sys.argv[1:]
    svg = OKLCH.sub(to_srgb, open(svg_path).read())
    left = OKLCH.findall(svg)
    if left:
        sys.exit(f"render-social: unconverted colours: {left}")
    with tempfile.NamedTemporaryFile("w", suffix=".svg") as srgb:
        srgb.write(svg)
        srgb.flush()
        result = subprocess.run(
            ["resvg", "--skip-system-fonts", "--use-fonts-dir", fonts, srgb.name, png_path],
            capture_output=True,
        )
    log = result.stderr.decode()
    if result.returncode != 0 or "Warning" in log:
        sys.exit(f"render-social: resvg failed or warned:\n{log}")
    print(f"render-social: {svg_path} -> {png_path}")


if __name__ == "__main__":
    main()
