import sys
import matplotlib.pyplot as plt
import numpy as np
from matplotlib import colors
from matplotlib.ticker import PercentFormatter
from matplotlib.lines import Line2D
import numpy as np
import seaborn as sns

def parse_input():
  try:
    time_metrics_file_name, inter_arrival_file_name  = sys.argv[1], sys.argv[2]
    frame_size, setup, conn_type = sys.argv[3], sys.argv[4], sys.argv[5]
  except:
    IndexError
    conn_type = None
    inter_arrival_file_name = None
    frame_size = None
    setup = None
    print("Usage: python3 ./PlotGenerator <filename1> <filename2> <framesize> <setup> <conn_type>")
    exit()

  return time_metrics_file_name, inter_arrival_file_name, frame_size, setup, conn_type

def parse_packet(packet_line):

  if len(packet_line) > 1:
    splitted_line = packet_line.split()
    packet_index = int(splitted_line[1])
    packet_end_to_end = int(splitted_line[6])
    packet_rtt = int(splitted_line[12])
    
    return packet_index, packet_end_to_end, packet_rtt
  
  else:
    return None, None, None

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

def get_setup(setup):
  match setup:
    case "lab":
      return "Server - lab, client - lab (Wired connection)"

    case "aroma":
      return "Server - lab , client - Aroma (Wifi connection)"
    
    case "home":
      return "Server - lab, client - Remote vpn connection"
    
def MicroToMilli(nums):
  return [num/1000.0 for num in nums]

def plot_histogram(packet_values: list, title: str, x_label: str, file_name: str, conn_type: str, setup: str) -> None:
  sns.set_theme(style="darkgrid")  
  normalized_values = MicroToMilli(packet_values)
  plt.figure(figsize=(10, 6))

  x, bins, p = plt.hist(normalized_values, bins=30, color='skyblue', edgecolor='black')

  plt.grid(visible=True, color='grey', linestyle='-.', linewidth=0.5, alpha=0.6)

  # Adding labels and title
  plt.xlabel(x_label, fontsize=14)
  plt.ylabel('Percentage [%]', fontsize=14)
  plt.suptitle(title, fontsize=18) 
  setup_string = get_setup(setup.strip()) + " [" + conn_type + "]"
  plt.title(setup_string, fontsize=15)

  # Normalize to percentage
  sum = 0
  for item in p: 
    sum += int(item.get_height())
  
  for item in p:
    item.set_height(100 * item.get_height() / sum) 
  
  mean = np.mean(normalized_values)
  mean_line = plt.axvline(mean, color='red', linestyle='dashed', linewidth=1)
  plt.text(mean, 90, f'{mean:.2f}', color='red', fontsize=9, ha='right')

  plt.ylim(0, 100)
  plt.xlim(0, 120)

  custom_lines = [Line2D([0], [0], color='red', linestyle='dashed', linewidth=1)]
  plt.legend(custom_lines, ['Mean'], loc='upper right')

  #plt.plot()
  plt.savefig(file_name+" "+frame_size, dpi=300)
  
def get_audio_length(frame_size):
  channels, sample_rate = 2, 48000
  second_to_milli = 1000

  result = ((frame_size /(sample_rate*channels) )* second_to_milli)
  if result.is_integer():
    return int(result)
  return result

def plot_graph(packet_values: list, title: str, x_label: str, y_label: str, file_name: str) -> None:
  normalized_values = np.copy(packet_values)

  for i in range(len(normalized_values)):
    normalized_values[i] /= 1000

  plt.figure(figsize=(10, 8))

  p = plt.plot(range(len(normalized_values)), normalized_values, color='skyblue')

  # Add x, y gridlines 
  plt.grid(visible = True, color ='grey', 
        linestyle ='-.', linewidth = 0.5, 
        alpha = 0.6) 

  # Adding labels and title
  plt.xlabel(x_label)
  plt.ylabel(y_label)
  plt.title(title, fontsize=18)
  
  #plt.show()
  plt.savefig(file_name+" "+frame_size, dpi=300)
  
  return  

if __name__=="__main__":
  
  time_metrics_file_name, inter_arrival_file_name, frame_size, setup, conn_type = parse_input()
  inter_arrival_file_name = inter_arrival_file_name.removesuffix(".txt")+ " "+str(frame_size)+".txt"
  time_metrics_file_name = time_metrics_file_name.removesuffix(".txt")+ " "+str(frame_size)+".txt"
  
  packets = parse_stats_file(time_metrics_file_name, "metrics")
  packet_indexes = [packets[i][0] for i in range(len(packets))]
  packet_end_to_ends = [packets[i][1] for i in range(len(packets))]
  packet_RTTs = [packets[i][2] for i in range(len(packets))]
 
  inter_arrivals = parse_stats_file(inter_arrival_file_name, "interArrival")
  plot_histogram(
      packet_values=packet_end_to_ends,
      title='End to End Latency: '+str(get_audio_length(int(frame_size))) + " millisecond frames",
      x_label='End to End [milliseconds]',
      file_name="./Plots/End To Ends/Packet End To Ends",
      setup=setup, conn_type=conn_type
      )
  
  plot_histogram(
    packet_values=inter_arrivals,
    title='Inter-Arrivals: '+str(get_audio_length(int(frame_size))) + " millisecond frames",
    x_label='Inter-Arrival Times [milliseconds]',
    file_name="./Plots/Inter Arrivals/Inter-Arrivals",
    setup=setup, conn_type=conn_type
    )
  
  plot_histogram(
    packet_values=packet_RTTs,
    title='Round Trip Time: '+str(get_audio_length(int(frame_size))) + " millisecond frames",
    x_label='Round Trip Time [milliseconds]',
    file_name="./Plots/RTTs/Round Trip Time",
    setup=setup, conn_type=conn_type
    )

