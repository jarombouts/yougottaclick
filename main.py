from flask import Flask, render_template, request

app = Flask(__name__)

# set template folder to /templstes
app.template_folder = './templates'


@app.route('/')
def index():
    print("rendering index.html")
    return render_template('index.html')


@app.route('/scroll', methods=['GET'])
def load_more():
    batch = request.args.get('batch', default=0, type=int)
    print(f"rendering load-more?batch={batch}")
    return render_template('scroll-next.html', batch=batch)


if __name__ == '__main__':
    app.run(debug=True)
