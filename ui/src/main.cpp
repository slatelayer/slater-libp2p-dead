#include <QProcess>
#include <QQmlContext>
#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QSslConfiguration>
#include <QFontDatabase>
#include <QQuickStyle>
#include <QtWebView>

class Core : public QObject {
    Q_OBJECT

    private:
    QProcess proc;
    
    public:
    Q_INVOKABLE void run(QString dir = "") {
        qDebug() << "run " << dir;

        QStringList args;
        args << dir;

        QString path = QDir(QCoreApplication::applicationDirPath()).filePath("core");

        proc.start(path, args);

        QObject::connect(&proc, SIGNAL(started()), this, SLOT(started()));
        QObject::connect(&proc, SIGNAL(finished(int,QProcess::ExitStatus)), this, SLOT(finished(int,QProcess::ExitStatus)));
        QObject::connect(&proc, SIGNAL(readyReadStandardOutput()), this, SLOT(readyReadStandardOutput()));
        QObject::connect(&proc, SIGNAL(readyReadStandardError()), this, SLOT(readyReadStandardError()));
        QObject::connect(&proc, SIGNAL(errorOccurred(QProcess::ProcessError)), this, SLOT(errorOccurred(QProcess::ProcessError)));
    }

    public slots:
    void started() {
        qDebug() << "started";
        emit start();
    }

    void finished(int exitCode, QProcess::ExitStatus exitStatus) {
        qDebug() << "finished " << exitCode;
        emit end(exitCode);
    }

    void readyReadStandardOutput() {
        qDebug() << "ready";

        QByteArray buf = proc.readAllStandardOutput();
        bool ok;
        int port = buf.toInt(&ok);
        if (!ok)
            port = -1;

        emit ready(port);
    }

    void readyReadStandardError() {
        QByteArray buf = proc.readAllStandardError();
        emit error(buf);
    }

    void errorOccurred(QProcess::ProcessError err) {
        emit error(proc.errorString());
    }

    signals:
    void start();
    void ready(int port);
    void end(int exitCode);
    void error(QString err);
};

int main(int argc, char *argv[]) {
    QSslConfiguration sslConf = QSslConfiguration::defaultConfiguration();
    sslConf.setPeerVerifyMode(QSslSocket::VerifyNone);
    QSslConfiguration::setDefaultConfiguration(sslConf);

    QGuiApplication app(argc, argv);
    QtWebView::initialize();

    app.setApplicationName("slater");
    app.setOrganizationName("slater");
    app.setOrganizationDomain("slater.local");

    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Regular.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Bold.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Black.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Light.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Italic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Medium.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-Thin.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-BlackItalic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-BoldItalic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-MediumItalic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-LightItalic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/Roboto-ThinItalic.ttf");
    QFontDatabase::addApplicationFont(":/res/fonts/RobotoMono-Regular.ttf");

    QQuickStyle::setStyle("Material");

    qmlRegisterSingletonType(QUrl("qrc:/slater/qml/Style.qml"), "Style", 1, 0, "Style");

    QQmlApplicationEngine engine;

    Core *core = new(Core);
    engine.rootContext()->setContextProperty("_core", core);

    engine.load("qrc:/slater/qml/main.qml");

    return app.exec();
}

#include "main.moc"
