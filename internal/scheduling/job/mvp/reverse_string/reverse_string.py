import click


@click.command()
@click.option('--input', default="INPUT 1")
def reverse_string(input):
  print(input[::-1])
  return (input[::-1])


if __name__ == '__main__':
  reverse_string()
