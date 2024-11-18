import matplotlib.pyplot as plt
import json
from csv import DictWriter

def sum_project_score(project_data: dict) -> float:
    project_score = 0
    requirements = project_data.get("requirements", [])
    for req in requirements:
        if req["completed"]:
            project_score += req["max_score"] / req["attempts"]
    return project_score

def get_total_scores(filename: str) -> dict:
    with open(filename) as f:
        data = json.load(f)

    # Initialize a dictionary to store total scores per model
    total_scores = {}

    # Calculate the total scores for the "url-shortener" project per model
    if "benchmarks" in data:
        data = data["benchmarks"][-1]
    for model, model_data in data["models"].items():
        total_scores[model] = {}
        for project, project_data in model_data["projects"].items():
            total_scores[model][project] = sum_project_score(project_data)

    with open("out/scores.json", "w") as f:
        json.dump(total_scores, f)
    
    return total_scores

def get_total_cost(filename: str) -> float:
    with open(filename) as f:
        data = json.load(f)
        
    cost = 0
    for model, model_data in data["models"].items():
        for project, project_data in model_data["projects"].items():
            for req in project_data["requirements"]:
                cost += req["cost"]

    return cost

def tab_simple(total_scores: dict, out_filename: str):
    # project/model claude gemini
    # p1            10     5
    # p2            8      6
    table = [["project/model"] + list(total_scores.keys())]
    model1 = list(total_scores.keys())[0]
    projects = list(total_scores[model1].keys())
    for project in projects:
        row = [project]
        for _, project_scores in total_scores.items():
            row.append(project_scores[project])
        table.append(row)

    print(table)

    with open(out_filename, "w") as f:
        writer = DictWriter(f, fieldnames=table[0])
        writer.writeheader()
        for row in table[1:]:
            writer.writerow(dict(zip(table[0], row))) 

if __name__ == "__main__":
    path = "out/2024-11-18T16:27:58+01:00"
    total_scores = get_total_scores(f"{path}/stats.json")
    tab_simple(total_scores, f"{path}/scores.csv")
    total_cost = get_total_cost(f'{path}/stats.json')
    print(f"total cost: ${total_cost} ({total_cost * 26} CzK)")