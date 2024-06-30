import pandas as pd
import re
import matplotlib.pyplot as plt

def parse_line(line):
    pattern = r"Frame size:\s+(\d+)\s+\|\s+Average RTT:\s+(\d+)\s+\|\s+Average Inter-Arrival:\s+(\d+)\s+\|\s+Jitter:\s+(\d+)"
    match = re.match(pattern, line)
    if match:
        frame_size, rtt, inter_arrival, jitter = match.groups()
        return {
            "Frame Size [milliseconds]": int(frame_size),
            "RTT [milliseconds]": int(rtt),
            "Inter-Arrival [milliseconds]": int(inter_arrival),
            "Jitter [milliseconds]": int(jitter)
        }
    return None

def create_table_plot(summarizedStatsFile, output_image="./Plots/Summarized Table.png"):
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

    plt.savefig(output_image, dpi=300)

if __name__ == "__main__":
    summarizedStatsFile = "./Stats/SummarizedStats.txt"  
    create_table_plot(summarizedStatsFile)
    print("TODO: ADD SUPPORT FOR 2.5 MSECS WINDOW SIZE")
