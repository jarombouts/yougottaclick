import random
import time
import networkx as nx
import matplotlib.pyplot as plt

# probability settings
prob_generate = 0.7  # base chance to generate subtasks
prob_num_subtasks = [0.4, 0.4, 0.2]  # probability distribution for 1, 2, or 3 subtasks
max_depth = 5  # maximum depth of recursion


def f(task: str, depth: int = 0, is_summarization_task: bool = False):
    print(f"{'  ' * depth}*  processing task: {task}")

    # base case: stop recursion if depth exceeds max_depth
    if (depth >= max_depth):
        answer = f"answer from {task} (max depth reached)"
        print(f"{'  ' * depth}<- returning: {answer}")
        return answer

    if is_summarization_task:
        answer = f"answer from {task} (summarization)"
        print(f"{'  ' * depth}<- returning: {answer}")
        return answer

    # decide whether to generate subtasks or return an answer
    if random.random() > prob_generate * (1 - (depth / max_depth)):
        # this simulates returning a direct answer. note that we adjust probability of generating subtasks
        # based on depth to avoid excessive recursion, just for the sake of the simulation
        answer = f"answer from {task}"
        print(f"{'  ' * depth}<- returning: {answer}")
        return answer
    else:
        # this simulates the creation of subtasks, in case the problem is too complex to solve in one go.
        num_subtasks = random.choices([1, 2, 3], prob_num_subtasks)[0]
        print(f"{'  ' * depth}-> generating {num_subtasks} subtasks for {task}")

        # create subtasks
        subtasks = [f"{task}.{i + 1}" for i in range(num_subtasks)]
        answers = []

        # recursively process subtasks
        for subtask in subtasks:
            answer = f(subtask, depth + 1)
            answers.append(answer)

        # summarize the answers retrieved from the subtasks
        summary = f"please summarize these {len(answers)} answers: {', '.join(answers)}"
        return f(task=summary, depth=-1, is_summarization_task=True)

f("T0")

# initial task
initial_task = "T0"

# run the recursive function
final_result = f(initial_task)

# display the final result
print("\nFinal result:", final_result)

# visualize the graph
pos = nx.multipartite_layout(G, subset_key='level', align='horizontal')
labels = {node: G.nodes[node]['label'] for node in G.nodes}
node_colors = ['#1f78b4' if G.nodes[node]['result'].startswith('summary') else '#33a02c' for node in G.nodes]

plt.figure(figsize=(12, 8))
nx.draw(G, pos, labels=labels, with_labels=True, node_size=1500, node_color=node_colors, font_size=10,
        font_color='white', font_weight='bold', arrowsize=15, arrowstyle='->')
plt.title('Recursive Task Execution Graph')
plt.show()
