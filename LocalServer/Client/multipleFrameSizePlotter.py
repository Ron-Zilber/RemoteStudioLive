import pandas as pd
import re
import matplotlib.pyplot as plt
import sys
import os

def parse_line(line):
    pattern = r"Frame size:\s+([\d.]+)\s+\|\s+Average End to End:\s+([\d.]+)\s+\|\s+Average RTT:\s+([\d.]+)\s+\|\s+Average Inter-Arrival:\s+([\d.]+)\s+\|\s+Jitter:\s+([\d.]+)\s+\|\s+Unordered Packets:\s*([\d.]+)%\s+\|\s+Lost Packets:\s*([\d.]+)%"
    match = re.match(pattern, line)
    if match:
        frame_size, end_to_end, rtt, inter_arrival, jitter, unordered_packets, lost_packets = match.groups()
        frame_size = float(frame_size)
        return {
            "Frame Size [milliseconds]": frame_size,
            "End to End [milliseconds]": float(end_to_end),
            "RTT [milliseconds]": float(rtt),
            "Jitter [milliseconds]": float(jitter),
            "Inter-Arrival [milliseconds]": float(inter_arrival),
            "Unordered Packets [%]": float(unordered_packets),
            "Lost Packets [%]": float(lost_packets)
        }
    return None

def create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png"):
    setupString = get_setup(setup.strip())

    try:
        with open(summarizedStatsFile, 'r') as file:
            lines = file.readlines()
    except FileNotFoundError:
        print(f"File not found: {summarizedStatsFile}")
        return

    data = [parse_line(line) for line in lines if parse_line(line) is not None]

    if not data:
        print("No data parsed from the file.")
        return

    df = pd.DataFrame(data)
    df = df[[
        "Frame Size [milliseconds]",
        "End to End [milliseconds]",
        "RTT [milliseconds]",
        "Jitter [milliseconds]",
        "Inter-Arrival [milliseconds]",
        "Unordered Packets [%]",
        "Lost Packets [%]"
    ]]

    _, ax = plt.subplots(figsize=(14, 6))  # Adjust size as needed
    ax.axis('tight')
    ax.axis('off')
    table = ax.table(cellText=df.values, colLabels=df.columns, cellLoc='center', loc='center')

    table.auto_set_font_size(True)
    #table.set_fontsize(10)
    table.scale(1.2, 1.2)

    # Make the first row bold
    cell_dict = table.get_celld()
    for i in range(len(df.columns)):
        cell_dict[(0, i)].set_text_props(fontweight='bold')

    plt.subplots_adjust(top=0.85)
    plt.suptitle("End-to-End, RTT, Inter-Arrival, Jitter, Unordered Packets, and Lost Packets Metrics for Various Frame Sizes", fontsize=14)
    plt.title("Setup: " + setupString)

    # Save the plot before showing it
    plt.savefig(output_image, dpi=300)

def parse_input():
    try:
        setup = sys.argv[1]
        return setup
    except IndexError:
        print("Usage: python3 ./multipleFrameSizePlotter.py <setup>")
        exit()

def get_setup(setup):
    match setup:
        case "lab":
            return "Server - lab, client - lab"
        case "aroma":
            return "Server - lab , client - Aroma"
        case "home":
            return "Server - lab, client - same city"
        case _:
            return "Unknown setup"

if __name__ == "__main__":
    setup = parse_input()
    summarizedStatsFile = "./Stats/SummarizedStats.txt"
    
    # Ensure the output directory exists
    os.makedirs(os.path.dirname("./Plots/Summarized Table.png"), exist_ok=True)
    
    create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png")
