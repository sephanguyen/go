# Intruction:
For evaluate mAP of omr machine learning model, we use from this [mAP tool](https://github.com/Cartucho/mAP)
This util predict bubble bbox then write to [Cartucho format](https://github.com/Cartucho/mAP#create-the-detection-results-files) for validation. 

# Usage:
```
Usage: bubble_validate.py [OPTIONS]

Options:
  --image_test TEXT  folder image which be use for validation
  --model_path TEXT  Model paths
  --help             Show this message and exit.

```

# Command:
```commandline
python bubble_validate.py
```

