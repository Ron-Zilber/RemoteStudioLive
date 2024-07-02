import pandas as pd
import re
import matplotlib.pyplot as plt
import sys

def parse_line(line):
    pattern = r"Frame size:\s+(\d+)\s+\|\s+Average RTT:\s+(\d+)\s+\|\s+Average Inter-Arrival:\s+(\d+)\s+\|\s+Jitter:\s+(\d+)"
    match = re.match(pattern, line)
    if match:
        frame_size, rtt, inter_arrival, jitter = match.groups()
        frame_size = float(frame_size)
        if frame_size == 2:
            frame_size = 2.5  # Change the specific frame size to 2.5
        else:
            frame_size = int(frame_size)  # Keep all other frame sizes as integers
        return {
            "Frame Size [milliseconds]": frame_size,
            "RTT [milliseconds]": int(rtt),
            "Inter-Arrival [milliseconds]": int(inter_arrival),
            "Jitter [milliseconds]": int(jitter)
        }
    return None

def create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png"):
    setupString = get_setup(setup.strip())
    with open(summarizedStatsFile, 'r') as file:
        lines = file.readlines()

    data = [parse_line(line) for line in lines if parse_line(line) is not None]

    if not data:
        return

    df = pd.DataFrame(data)

    fig, ax = plt.subplots(figsize=(10, 2))  # Adjust size as needed
    ax.axis('tight')
    ax.axis('off')
    table = ax.table(cellText=df.values, colLabels=df.columns, cellLoc='center', loc='center')

    table.auto_set_font_size(False)
    table.set_fontsize(10)
    table.scale(1.2, 1.2)

    # Make the first row bold
    cell_dict = table.get_celld()
    for i in range(len(df.columns)):
        cell_dict[(0, i)].set_text_props(fontweight='bold')

    plt.subplots_adjust(top=0.75)
    plt.suptitle("Summary of RTT, Inter-Arrival, and Jitter Metrics for Various Frame Sizes", fontsize=14)
    plt.title("Setup: " + setupString)
    #plt.show()
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
    create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png")
    print("TODO: ADD SUPPORT FOR 2.5 MSECS WINDOW SIZE")
