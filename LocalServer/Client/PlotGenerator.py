import sys
import matplotlib.pyplot as plt
import numpy as np
from matplotlib import colors
from matplotlib.ticker import PercentFormatter
import numpy as np


def check_input_file():
  try:
    rtt_file_name = sys.argv[1]
    inter_arrival_file_name = sys.argv[2]
  except:
    IndexError
    stats_file_name = None
    inter_arrival_file_name = None
    print("Usage: python3 ./PlotGenerator <filename1> <filename2>")
    exit()

  return rtt_file_name, inter_arrival_file_name

def parse_packet(packet_line):
  if len(packet_line) > 1:
    splitted_line = packet_line.split()
    #print("Splitted words are: ", splitted_line)
    packet_index = int(splitted_line[1])
    packet_RTT = int(splitted_line[6])
    return packet_index, packet_RTT
  
  else:
    return None, None

def parse_stats_file(stats_file_name, type):
    statsFile = open(stats_file_name)

    result = []

    match type:
      case "rtt":
        for line in statsFile:
          packet_index, packet_RTT = parse_packet(line)
          if packet_index == None:
            break
          else:
            result.append((packet_index, packet_RTT))

      case "interArrival":
        for num in statsFile:
          if num != "\n":
            result.append(int(num))
            #print(int(num))
      
    statsFile.close()
    return result
  
def plot_histogram(packet_values: list, title: str, x_label: str, file_name: str) -> None:
  normalized_values = np.copy(packet_values)
  plt.figure(figsize=(10, 6))
  # Plotting a basic histogram
  for i in range(len(normalized_values)):
    normalized_values[i] /= 1000

  x, bins, p = plt.hist(normalized_values, bins=60, color='skyblue', edgecolor='black')

  # Add x, y gridlines 
  plt.grid(visible = True, color ='grey', 
        linestyle ='-.', linewidth = 0.5, 
        alpha = 0.6) 

  # Adding labels and title
  plt.xlabel(x_label)
  plt.ylabel('Percentage [%]')
  plt.title(title)
  

  # Normalize to percentage
  sum = 0
  for item in p: 
    sum += int(item.get_height())
  
  for item in p:
    item.set_height(100 * item.get_height() / sum)   


  plt.ylim(0, 100)
  #plt.xlim(0, 100)

  #plt.plot()
  plt.savefig(file_name, dpi=300)
  

def plot_graph(packet_values: list, title: str, x_label: str, y_label: str, file_name: str) -> None:
  normalized_values = np.copy(packet_values)

  for i in range(len(normalized_values)):
    normalized_values[i] /= 1000

  plt.figure(figsize=(10, 6))


  p = plt.plot(range(len(normalized_values)), normalized_values, color='skyblue')

  # Add x, y gridlines 
  plt.grid(visible = True, color ='grey', 
        linestyle ='-.', linewidth = 0.5, 
        alpha = 0.6) 

  # Adding labels and title
  plt.xlabel(x_label)
  plt.ylabel(y_label)
  plt.title(title)
  
  #plt.show()
  plt.savefig(file_name, dpi=300)
  
  return  

if __name__=="__main__":
  
  rtt_file_name, inter_arrival_file_name = check_input_file()
  packets = parse_stats_file(rtt_file_name, "rtt")
  packet_indexes = [packets[i][0] for i in range(len(packets))]
  packet_RTTs = [packets[i][1] for i in range(len(packets))]

  inter_arrivals = parse_stats_file(inter_arrival_file_name, "interArrival")

  plot_graph(packet_RTTs, "Packets Round Trip Time", "Packet Index", "RTT [milliseconds]", "./Plots/test")
  
  plot_histogram(packet_RTTs, 'Packets Round Trip Time (RTT)',
                 'RTT [milliseconds]', "./Plots/Packet RTTs")
  
  plot_histogram(inter_arrivals, 'Packet Inter-Arrival Times',
                 'Inter-Arrival Times [milliseconds]', "./Plots/Inter-Arrivals")






