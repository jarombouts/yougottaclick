import random
import networkx as nx
from pydantic import BaseModel
from typing import List, Optional
import json
import matplotlib.pyplot as plt


# probability settings for mock subtask generation
prob_generate = 0.7
prob_num_subtasks = [0.4, 0.4, 0.2]


class TaskDescription(BaseModel):
    task_id: str
    task_name: str
    depth: int = 0
    max_depth: int


class TaskResult(BaseModel):
    task_id: str
    task_name: str
    depth: int
    result: str
    subtasks: Optional[List['TaskResult']] = None


def F(task_description: TaskDescription) -> nx.DiGraph:
    # create a new directed graph for this task
    G = nx.DiGraph()

    # add the current task as a node in the graph
    G.add_node(task_description.task_id, label=task_description.task_name, depth=task_description.depth)

    # check if maximum depth is reached, if so, this node becomes a leaf node by default
    if task_description.depth >= task_description.max_depth:
        # todo replace by function that generates result without allowing subtasks
        #   result = call_llm_no_subtasks(task_description)
        result = f"answer from {task_description.task_name} (max depth reached)"
        G.nodes[task_description.task_id]['result'] = result
        return G

    # todo add function to generate result with optional subtasks
    #   result = call_llm_with_subtasks(task_description)

    # else, determine if subtasks should be executed
    if should_execute_subtasks(task_description):
        G = execute_subtasks(G, task_description)
    # finally, if no subtasks are executed, this node becomes a leaf node
    else:
        result = f"answer from {task_description.task_name}"
        G.nodes[task_description.task_id]['result'] = result

    return G


def execute_subtasks(G, task_description):
    num_subtasks = determine_num_subtasks()
    subtasks = []
    for i in range(num_subtasks):
        subtask_id = f"{task_description.task_id}.{i + 1}"
        subtask_name = f"{task_description.task_name}.{i + 1}"

        # create subtask description
        subtask_description = TaskDescription(
            task_id=subtask_id,
            task_name=subtask_name,
            depth=task_description.depth + 1,
            max_depth=task_description.max_depth
        )

        # recursive call to F() for the subtask
        subtask_result = F(subtask_description)

        # merge the subtask graph into the current graph
        G = nx.compose(G, subtask_result)

        # add an edge from the current task to the subtask
        G.add_edge(task_description.task_id, subtask_id)

        subtasks.append(subtask_result)
    # summarize the answers
    summary = f"summary of {len(subtasks)} subtasks"
    G.nodes[task_description.task_id]['result'] = summary
    return G


def should_execute_subtasks(task_description: TaskDescription) -> bool:
    # implement logic or placeholder logic for subtask generation
    return random.random() < prob_generate


def determine_num_subtasks() -> int:
    # logic to determine the number of subtasks
    return random.choices([1, 2, 3], prob_num_subtasks)[0]


############################################################################################################

# create an initial task description
initial_task = TaskDescription(task_id="T0", task_name="T0", max_depth=5)

# run the recursive function
final_graph = F(initial_task)

# visualize the graph

def graph_to_json(G: nx.DiGraph) -> str:
    from networkx.readwrite import json_graph
    return json.dumps(json_graph.node_link_data(G))


def json_to_graph(data: str) -> nx.DiGraph:
    from networkx.readwrite import json_graph
    return json_graph.node_link_graph(json.loads(data))


pos = nx.multipartite_layout(final_graph, subset_key='depth', align='horizontal')
labels = {node: final_graph.nodes[node]['label'] for node in final_graph.nodes}
node_colors = ['#1f78b4' if final_graph.nodes[node]['result'].startswith('summary') else '#33a02c' for node in
               final_graph.nodes]

plt.figure(figsize=(12, 8))
nx.draw(final_graph, pos, labels=labels, with_labels=True, node_size=1500, node_color=node_colors, font_size=10,
        font_color='black', font_weight='bold', arrowsize=15, arrowstyle='->')
plt.title('Recursive Task Execution Graph')
plt.show()
