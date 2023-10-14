from internal.aphelios.model.omr.python.model.utils.utils import loading_data


def coco2dota(images, segments, labels, categories, root_output):
  for i in range(0, len(images)):
    filename = images[i].split("/")[-1]
    filename = filename.split(".")[0]
    with open(f"{root_output}/{filename}.txt", "w") as file:
      line = ""

      # 1. A[0:8]: Polygons with format (x1, y1, x2, y2, x3, y3, x4, y4)
      for j in range(8):
        line = line + f"{segments[i][j]}" + " "

      # 2. A[8]: Category.
      line = line + f"{labels[i]}" + " "

      # 3. A[9]: Difficulty.
      line = line + "0"

      file.write(line)
      print(f"{filename}.txt - {line}")


if __name__ == '__main__':
  data_path = "result.json"
  root_output = "./dota_format"

  #1. Data preprocess (formating)
  images, segments, labels, categories = loading_data(data_path)
  dota_fmt = coco2dota(images, segments, labels, categories, root_output)

