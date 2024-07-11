import pandas as pd
import matplotlib.pyplot as plt
import sys, os, re, PlotGenerator
import numpy as np
import seaborn as sns
from matplotlib.lines import Line2D

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
        summarized_inter_arrivals.append(PlotGenerator.MicroToMilli(inter_arrivals))

    return summarized_end_to_ends, summarized_RTTs, summarized_inter_arrivals

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

def parse_input():
    try:
        setup = sys.argv[1]
        conn_type = sys.argv[2]
        return setup, conn_type
    except IndexError:
        print("Usage: python3 ./multipleFrameSizePlotter.py <setup> <conn_type>")
        exit()

def parse_stats_file(stats_file_name, type):
    statsFile = open(stats_file_name)

    result = []

    match type:
      case "metrics":
        for line in statsFile:
          packet_index, packet_end_to_end, packet_rtt = PlotGenerator.parse_packet(line)
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

def bins_to_percentage(p):
   sum = 0  
   for item in p:
      sum += int(item.get_height())

   for item in p:
      item.set_height(100 * item.get_height() / sum)

def create_table_plot(summarizedStatsFile, setup, conn_type, output_image="./Plots/Summarized Table.png"):
    setupString = PlotGenerator.get_setup(setup.strip())
    data = parse_data(summarizedStatsFile)

    df = pd.DataFrame(data)
    df = df[["Frame Size [milliseconds]", "End to End [milliseconds]",
        "RTT [milliseconds]", "Jitter [milliseconds]", 
        "Inter-Arrival [milliseconds]", "Unordered Packets [%]",
        "Lost Packets [%]"]]

    fig, ax = plt.subplots(figsize=(12,2.6))  # Adjust size as needed
    ax.axis('tight')
    ax.axis('off')
    table = ax.table(cellText=df.values, colLabels=df.columns, cellLoc='center', loc='center')

    table.auto_set_font_size(True)
    #table.set_fontsize(10)
    table.scale(1.2, 1.2)

    fig.tight_layout()

    # Make the first row bold
    cell_dict = table.get_celld()
    for i in range(len(df.columns)):
        cell_dict[(0, i)].set_text_props(fontweight='bold', color='black')
        cell_dict[(0, i)].set_facecolor('skyblue')

    for row in range(1, len(df) + 1):
        for col in range(len(df.columns)):
            if row % 2 == 0:
                cell_dict[(row, col)].set_facecolor('#f1f1f2')
            else:
                cell_dict[(row, col)].set_facecolor('#e2e2e2')

    for key, cell in cell_dict.items():
        cell.set_edgecolor('grey')
        cell.set_linewidth(0.5)
        cell.set_text_props(ha='center', va='center', fontsize=14)

    plt.subplots_adjust(top=0.65)
    plt.suptitle("End-to-End, RTT, Inter-Arrival, Jitter, Unordered Packets,\nand Lost Packets Metrics for Various Frame Sizes", fontsize=17.5)
    plt.title("Setup: " + setupString + " ["+conn_type+"]", fontsize=14.5)

    # Save the plot before showing it
    plt.savefig(output_image, dpi=300)
    plt.close()

def create_box_plot(data, y_limit, suptitle, x_label, y_label, box_colors, setup, conn_type, output_file="box_plot.png"):
    sns.set_theme(style="darkgrid")  # Set the Seaborn style directly

    fig, ax = plt.subplots(figsize=(10, 7))
    
    box = ax.boxplot(data, patch_artist=True, medianprops=dict(color="red"))
    
    # Set box colors
    for i in range(len(box['boxes'])):
        box['boxes'][i].set_facecolor(box_colors[i])
        box['boxes'][i].set_alpha(0.5)
    
    # Customize grid
    ax.yaxis.grid(True, linestyle='-.', which='major', color='lightgrey', alpha=0.5)

    # Set the axes ranges and axes labels
    x_labels = [1.25, 2.5, 5, 10, 20]
    num_boxes = len(data)
    ax.set_xlim(0.5, num_boxes + 0.5)
    ax.set_ylim(bottom=-1, top=y_limit)  # Set this as per your data range

    # Set custom labels
    ax.set_xticklabels(x_labels, rotation=0, fontsize=10)
    ax.set_xlabel(x_label, fontsize=14)
    ax.set_ylabel(y_label, fontsize=14)

    # Annotate the median values
    for i in range(num_boxes):
        median = box['medians'][i].get_ydata()[0]
        ax.annotate(f'{median:.2f}', xy=(i + 1, median), xytext=(0, 5), 
                    textcoords='offset points', ha='right', fontsize=10, color='red')

        # Calculate and plot the mean
        mean = np.mean(data[i])
        ax.plot([i + 0.75, i + 1.25], [mean, mean], linestyle='--', color='blue')
        ax.annotate(f'{mean:.2f}', xy=(i + 1, mean), xytext=(0, -15), textcoords='offset points', 
                    ha='left', fontsize=10, color='blue')

    # Add a legend
    custom_lines = [Line2D([0], [0], color='red', lw=2, label='Median'),
        Line2D([0], [0], color='blue', lw=2, linestyle='--', label='Mean')]
    ax.legend(handles=custom_lines, loc='upper right')

    setup_string = PlotGenerator.get_setup(setup.strip())
    subtitle = setup_string + " [" + conn_type + "]"

    # Set titles
    plt.suptitle(suptitle, fontsize=18)
    ax.set_title(subtitle, fontsize=15)

    # Save the plot as an image file
    plt.savefig(output_file, bbox_inches='tight', dpi=300)
    plt.close()
    return
def create_summarized_box_plots(data, y_limits, y_labels_list, titles_list,setup, conn_type, output_files_list):
    colors = ['skyblue', 'orange', 'green', 'red', 'purple']

    for i in range(len(data)):
      create_box_plot(data[i],suptitle= titles_list[i],setup=setup,conn_type=conn_type,box_colors=colors,
         x_label= "Frame Size [millisecond]",y_label= y_labels_list[i],
         output_file= output_files_list[i], y_limit= y_limits[i])
    
    return

def create_summarized_histograms(metrics_files, inter_arrival_files, x_labels, r_widths, num_bins, x_limits, titles, setup, conn_type, output_file_names=None):
    sns.set_theme(style="darkgrid")

    summarized_RTTs = []
    summarized_end_to_ends = []
    summarized_inter_arrivals = []

    for i in range(len(metrics_files)):
        packets = parse_stats_file(metrics_files[i], "metrics")
        #packet_indexes = [packets[i][0] for i in range(len(packets))]
        end_to_ends = [packets[i][1] for i in range(len(packets))]
        rtts = [packets[i][2] for i in range(len(packets))]
        inter_arrivals = parse_stats_file(inter_arrival_files[i], "interArrival")

        summarized_end_to_ends.append(PlotGenerator.MicroToMilli(end_to_ends))
        summarized_RTTs.append(PlotGenerator.MicroToMilli(rtts))
        summarized_inter_arrivals.append(PlotGenerator.MicroToMilli(inter_arrivals))
        summarized_metrics = [summarized_end_to_ends, summarized_RTTs, summarized_inter_arrivals]
        
    colors = ['skyblue', 'orange', 'green', 'red', 'purple']

    for i in range(len(summarized_metrics)):
        plt.figure(figsize=(10,6))
        for j in range(len(summarized_end_to_ends)):
            x, b, p = plt.hist(summarized_metrics[i][j],bins=num_bins[i], histtype='bar', rwidth=r_widths[i], color=colors[j], edgecolor='black', alpha=0.5)
            bins_to_percentage(p)
            mean = np.mean(summarized_metrics[i][j])
            mean_line = plt.axvline(mean, color=colors[j], linestyle='dashed', linewidth=1)
            max_height = max(item.get_height() for item in p)
            plt.text(mean, min(max_height + 2, 95), f'{mean:.2f}', color=colors[j], fontsize=10, ha='right')
        
        plt.plot()

        plt.grid(visible=True, color='lightgrey', linestyle='-.', linewidth=0.5, alpha=0.6)
        # Add a legend
        labels = ['Frame size: 1.25  mS', 'Frame size:   2.5  mS', 'Frame size:      5  mS', 'Frame size:    10  mS', 'Frame size:    20  mS']

        custom_lines = [ Line2D([0], [0], color=colors[i], lw=2, label=labels[i]) for i in range(5)]
 
        plt.legend(handles=custom_lines, loc='upper right')

        plt.ylabel("Percentage [%]", fontsize=14)
        plt.xlabel(x_labels[i], fontsize=14)
        plt.suptitle(titles[i], fontsize=18)
        setup_string = PlotGenerator.get_setup(setup.strip()) + " ["+conn_type+"]"
        plt.title(setup_string, fontsize=15)

        plt.ylim(0, 100)
        plt.xlim(0, x_limits[i])
        plt.savefig(output_file_names[i], dpi=300)
        plt.close()

if __name__ == "__main__":

    setup, conn_type = parse_input()
    summarized_stats_file = "./Stats/SummarizedStats.txt"
    
    # Ensure the output directory exists
    os.makedirs(os.path.dirname("./Plots/Summarized Table.png"), exist_ok=True)
    
    create_table_plot(summarized_stats_file, setup, conn_type, output_image="./Plots/Summarized Table.png")

    data = load_data()

    titles = ["End To End Latency","Round Trip Time",
       "Inter-Arrivals"]
    
    y_labels = ["End To End Latency [millisecond]","Round Trip Time [millisecond]",
       "Inter-Arrivals [millisecond]"]
    
    output_files = ["./Plots/Summarized End to End - Box Plot.png","./Plots/Summarized Inter Arrival - Box Plot.png",
       "./Plots/Summarized RTT - Box Plot.png"]

    y_limits=[60, 10, 60]

    create_summarized_box_plots(data, y_limits, y_labels, titles, setup, conn_type, output_files)

    metrics_files = ["./Stats/StatisticsLog 120.txt", "./Stats/StatisticsLog 240.txt",
       "./Stats/StatisticsLog 480.txt","./Stats/StatisticsLog 960.txt",
       "./Stats/StatisticsLog 1920.txt"]
    
    inter_arrival_files = ["./Stats/interArrivalLog 120.txt", "./Stats/interArrivalLog 240.txt",
       "./Stats/interArrivalLog 480.txt", "./Stats/interArrivalLog 960.txt",
       "./Stats/interArrivalLog 1920.txt"]

    x_labels = ["End To End Latency [millisecond]", "Round Trip Time [millisecond]", "Packet Inter-Arrivals [millisecond]"]

    titles = ["End to End Latency", "Round Trip Times", "Inter Arrivals"]

    x_limits = [60, 10, 60]
    r_widths = [0.6, 0.1, 0.6]
    num_bins = [120, 30, 120]

    output_files = ["./Plots/Summarized End to End - Histogram.png", "./Plots/Summarized RTT - Histogram.png",
                    "./Plots/Summarized Inter Arrival - Histogram.png"]

    create_summarized_histograms(inter_arrival_files=inter_arrival_files,
                               metrics_files=metrics_files,
                               x_labels=x_labels, x_limits=x_limits, num_bins=num_bins,
                               r_widths=r_widths, titles=titles,conn_type=conn_type,
                               setup=setup, output_file_names=output_files)
