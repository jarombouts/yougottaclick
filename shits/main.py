from flask import Flask, render_template, request

app = Flask(__name__)

# set template folder to /templstes
app.template_folder = './templates'


@app.route('/')
def index():
    print("rendering index.html")
    return render_template('index.html')


@app.route('/scroll', methods=['GET'])
def scroll():
    batch = request.args.get('batch', default=0, type=int) % 1024
    up = request.args.get('up', default=False, type=bool)
    if up:
        return render_template('scroll-up.html', batch=batch)
    else:
        return render_template('scroll.html', batch=batch)


if __name__ == '__main__':
    app.run(debug=True)
