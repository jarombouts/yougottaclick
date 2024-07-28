from flask import Flask, render_template, request

app = Flask(__name__)

# set template folder to /templstes
app.template_folder = './templates'


@app.route('/')
def index():
    print("rendering index.html")
    return render_template('index.html')


@app.route('/load-more', methods=['GET'])
def load_more():
    batch = request.args.get('batch', default=0, type=int)
    prev = request.args.get('prev', default=False, type=bool)

    print(f"rendering load-more?batch={batch}&prev={prev}")

    if prev:
        return render_template('load-prev.html', batch=batch)
    else:
        return render_template('load-next.html', batch=batch)



if __name__ == '__main__':
    app.run(debug=True)
