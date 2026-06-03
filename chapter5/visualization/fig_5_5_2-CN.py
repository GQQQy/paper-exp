#!/usr/bin/env python3
"""Compatibility wrapper for Chapter 5 Figure 24/25 generation.

The actual data pipeline lives in chapter5_all_figures.py and reads
experiment/logs/raw_experiment_log.json instead of hard-coded plot arrays.
"""

from chapter5_all_figures import main


if __name__ == "__main__":
    main()
