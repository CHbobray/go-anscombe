"""Whole-program benchmark: Miller's Python and R programs versus Go.

Runs each complete program many times as a separate process, the way a user
would run it, and reports wall-clock time and peak memory per run. This
measures the entire program (interpreter start-up, loading packages, fitting
all four Anscombe sets and printing the summaries), not just the regression.

Programs compared:
  Python  python/miller_anscombe.py   Miller (2015) Exhibit 1.1, plotting removed
  R       r/miller_anscombe.R         Miller (2015) Exhibit 1.2, plotting removed
  Go      anscombe -reps 0 -mc 0      the Go application, summaries only

Usage (from the repository root):
  python3 benchmark/compare_programs.py [runs]

Peak memory needs os.wait4, which exists on macOS and Linux. On Windows the
script still reports times and leaves the memory column blank.
"""

import os
import platform
import statistics
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


def go_binary() -> Path:
    """Pick the prebuilt executable for this computer, or build one."""
    system = platform.system().lower()
    machine = platform.machine().lower()
    arch = "arm64" if machine in ("arm64", "aarch64") else "amd64"
    name = {"windows": "anscombe-windows-amd64.exe",
            "darwin": f"anscombe-darwin-{arch}",
            "linux": "anscombe-linux-amd64"}.get(system)
    if name and (ROOT / "bin" / name).exists():
        path = ROOT / "bin" / name
        if system != "windows":
            path.chmod(0o755)  # web downloads can drop the executable bit
        return path
    built = ROOT / "bin" / "anscombe-local"
    subprocess.run(["go", "build", "-o", str(built), "./cmd/anscombe"], cwd=ROOT, check=True)
    return built


def run_once(command: list[str]) -> tuple[float, float | None]:
    """Run a command once, returning seconds elapsed and peak memory in MB."""
    start = time.perf_counter()
    process = subprocess.Popen(command, cwd=ROOT, stdout=subprocess.DEVNULL,
                               stderr=subprocess.PIPE)
    if hasattr(os, "wait4"):
        _, status, usage = os.wait4(process.pid, 0)
        elapsed = time.perf_counter() - start
        process.returncode = os.waitstatus_to_exitcode(status)
        scale = 1e6 if sys.platform == "darwin" else 1e3  # bytes on macOS, KB on Linux
        peak_mb = usage.ru_maxrss / scale
    else:
        process.wait()
        elapsed = time.perf_counter() - start
        peak_mb = None
    if process.returncode != 0:
        error = process.stderr.read().decode(errors="replace")
        raise RuntimeError(f"{' '.join(command)} failed:\n{error}")
    return elapsed, peak_mb


def main() -> None:
    runs = int(sys.argv[1]) if len(sys.argv) > 1 else 20
    if runs < 1:
        sys.exit("runs must be at least 1")

    programs = {
        "Go": [str(go_binary()), "-reps", "0", "-mc", "0"],
        "Python (Miller)": [sys.executable, "python/miller_anscombe.py"],
        "R (Miller)": ["Rscript", "r/miller_anscombe.R"],
    }

    print(f"Whole-program benchmark: {runs} runs of each program "
          f"({platform.system()} {platform.machine()})\n")
    print(f"{'Program':<17} {'Mean s':>9} {'SD s':>8} {'Min s':>8} {'Peak MB':>9}")
    for name, command in programs.items():
        run_once(command)  # warm-up run so file caches do not favour any program
        times, peaks = [], []
        for _ in range(runs):
            elapsed, peak = run_once(command)
            times.append(elapsed)
            peaks.append(peak)
        sd = statistics.stdev(times) if runs > 1 else 0.0
        peak_text = f"{max(peaks):9.1f}" if peaks[0] is not None else f"{'n/a':>9}"
        print(f"{name:<17} {statistics.mean(times):9.4f} {sd:8.4f} {min(times):8.4f} {peak_text}")


if __name__ == "__main__":
    main()
