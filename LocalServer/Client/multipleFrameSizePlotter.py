import pandas as pd
import re
import matplotlib.pyplot as plt
import sys
import os
import numpy as np
import seaborn as sns
from matplotlib.lines import Line2D

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

def parse_data(data_file):
    try:
        with open(data_file, 'r') as file:
            lines = file.readlines()
    except FileNotFoundError:
        print(f"File not found: {data_file}")
        return

    data = [parse_line(line) for line in lines if parse_line(line) is not None]

    if not data:
        print("No data parsed from the file.")
        return
    return data

def create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png"):
    setupString = get_setup(setup.strip())

    data = parse_data(summarizedStatsFile)
    

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

def parse_packet(packet_line):

  if len(packet_line) > 1:
    splitted_line = packet_line.split()
    packet_index = int(splitted_line[1])
    packet_end_to_end = int(splitted_line[6])
    packet_rtt = int(splitted_line[12])
    
    return packet_index, packet_end_to_end, packet_rtt
  
  else:
    return None, None, None

def load_data():
    files_list = [["./Stats/interArrivalLog 120.txt",   "./Stats/StatisticsLog 120.txt"],
                  ["./Stats/interArrivalLog 240.txt",   "./Stats/StatisticsLog 240.txt"],
                  ["./Stats/interArrivalLog 480.txt",   "./Stats/StatisticsLog 480.txt"],
                  ["./Stats/interArrivalLog 960.txt",   "./Stats/StatisticsLog 960.txt"],
                  ["./Stats/interArrivalLog 1920.txt", "./Stats/StatisticsLog 1920.txt"]]
    summarized_RTTs = []
    summarized_inter_arrivals = []
    summarized_end_to_ends = []

    for pair in files_list:
        inter_arrival_file_name = pair[0]
        time_metrics_file_name  = pair[1]

        packets = parse_stats_file(time_metrics_file_name, "metrics")
        # packet_indexes = [packets[i][0] for i in range(len(packets))]
        packet_end_to_ends = [packets[i][1]/1000.0 for i in range(len(packets))]
        packet_RTTs = [packets[i][2]/1000.0 for i in range(len(packets))]
        inter_arrivals = parse_stats_file(inter_arrival_file_name, "interArrival")

        summarized_end_to_ends.append(packet_end_to_ends)
        summarized_RTTs.append(packet_RTTs)
        summarized_inter_arrivals.append(MicroToMilli(inter_arrivals))

    return summarized_end_to_ends, summarized_RTTs, summarized_inter_arrivals

def MicroToMilli(nums):
  return [num/1000 for num in nums]

def parse_stats_file(stats_file_name, type):
    statsFile = open(stats_file_name)

    result = []

    match type:
      case "metrics":
        for line in statsFile:
          packet_index, packet_end_to_end, packet_rtt = parse_packet(line)
          if packet_index == None:
            break
          else:
            result.append((packet_index, packet_end_to_end, packet_rtt))

      case "interArrival":
        for num in statsFile:
          if num != "\n":
            result.append(int(num))
      
    statsFile.close()
    return result

def create_box_plot(data, y_limit, title="Box Plot", x_label="Frame Size", y_label="Metric Value", output_file="box_plot.png", box_color="skyblue"):
    sns.set_theme(style="darkgrid")  # Set the Seaborn style directly

    fig, ax = plt.subplots(figsize=(10, 7))
    
    box = ax.boxplot(data, patch_artist=True, medianprops=dict(color="red"))
    
    # Set box colors
    for patch in box['boxes']:
        patch.set_facecolor(box_color)
    
    # Customize grid
    ax.yaxis.grid(True, linestyle='-', which='major', color='lightgrey', alpha=0.5)

    # Set the axes ranges and axes labels
    x_labels = [1.25, 2.5, 5, 10, 20]
    num_boxes = len(data)
    ax.set_xlim(0.5, num_boxes + 0.5)
    ax.set_ylim(bottom=-1, top=y_limit)  # Set this as per your data range

    # Set custom labels
    ax.set_xticklabels(x_labels, rotation=0, fontsize=10)
    ax.set_title(title, fontsize=18)
    ax.set_xlabel(x_label, fontsize=14)
    ax.set_ylabel(y_label, fontsize=14)

    # Annotate the median values
    for i in range(num_boxes):
        median = box['medians'][i].get_ydata()[0]
        ax.annotate(f'{median:.2f}', 
                    xy=(i + 1, median), 
                    xytext=(0, 5), 
                    textcoords='offset points', 
                    ha='center', 
                    fontsize=10, 
                    color='red')

        # Calculate and plot the mean
        mean = np.mean(data[i])
        ax.plot([i + 0.75, i + 1.25], [mean, mean], linestyle='--', color='blue')
        ax.annotate(f'{mean:.2f}', 
                    xy=(i + 1, mean), 
                    xytext=(0, -15), 
                    textcoords='offset points', 
                    ha='center', 
                    fontsize=10, 
                    color='blue')

    # Add a legend
    custom_lines = [
        Line2D([0], [0], color='red', lw=2, label='Median'),
        Line2D([0], [0], color='blue', lw=2, linestyle='--', label='Mean')
    ]
    ax.legend(handles=custom_lines, loc='upper right')

    # Save the plot as an image file
    plt.savefig(output_file, bbox_inches='tight', dpi=300)
    plt.close()
    return

def create_summarized_box_plots(data, y_limits, y_labels_list, titles_list,output_files_list):

   for i in range(len(data)):
      create_box_plot(
         data[i],
         title= titles_list[i],
         x_label= "Frame Size [millisecond]",
         y_label= y_labels_list[i],
         output_file= output_files_list[i],
         y_limit= y_limits[i]
         )
   return


if __name__ == "__main__":


    setup = parse_input()
    summarizedStatsFile = "./Stats/SummarizedStats.txt"
    
    # Ensure the output directory exists
    os.makedirs(os.path.dirname("./Plots/Summarized Table.png"), exist_ok=True)
    
    create_table_plot(summarizedStatsFile, setup, output_image="./Plots/Summarized Table.png")

    data = load_data()

    titles = [
       "End To End Latency [millisecond]", 
       "Round Trip Time [millisecond]",
       "Packet Inter-Arrivals [millisecond]"
       ]
    
    y_labels = [
       "End To End Latency [millisecond]", 
       "Round Trip Time [millisecond]",
       "Packet Inter-Arrivals [millisecond]"
       ]
    
    output_files = [
       "./Plots/Summarized End to End.png",
       "./Plots/Summarized Inter Arrival.png",
       "./Plots/Summarized RTT.png"
    ]
    y_limits=[60, 10, 60]
    create_summarized_box_plots(data, y_limits, y_labels, titles, output_files)

    