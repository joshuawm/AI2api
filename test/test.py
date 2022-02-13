import argparse

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--env",default="envFail")
    parser.add_argument("--from",default="fromFail")
    args = parser.parse_args()
    print(args)