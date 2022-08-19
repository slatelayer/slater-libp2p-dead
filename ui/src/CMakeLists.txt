cmake_minimum_required(VERSION 3.16)

project(slater VERSION 0.1 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 COMPONENTS Core Quick QuickControls2 REQUIRED)
find_package(Qt6 REQUIRED COMPONENTS WebView)

set(SOURCES main.cpp)

qt_add_resources(SOURCES res.qrc)

qt_add_executable(slater
    ${SOURCES}
)

set_source_files_properties(qml/Elements.qml PROPERTIES
    QT_QML_SINGLETON_TYPE TRUE
)

set_source_files_properties(qml/Style.qml PROPERTIES
    QT_QML_SINGLETON_TYPE TRUE
)

qt_add_qml_module(slater
    URI slater
    VERSION 1.0
    QML_FILES
        qml/main.qml
        qml/Bridge.qml
        qml/View.qml
        qml/Slate.qml
        qml/Style.qml
        qml/Element.qml
        
        qml/CoreText.qml
        qml/CoreSecretText.qml
        qml/CoreWeb.qml

        qml/InputElement.qml
        qml/CoreTextInput.qml
        qml/CoreChoiceInput.qml
        qml/CoreSecretTextInput.qml
)

set_target_properties(slater PROPERTIES
    MACOSX_BUNDLE_GUI_IDENTIFIER slater
    MACOSX_BUNDLE_BUNDLE_VERSION ${PROJECT_VERSION}
    MACOSX_BUNDLE_SHORT_VERSION_STRING ${PROJECT_VERSION_MAJOR}.${PROJECT_VERSION_MINOR}
    MACOSX_BUNDLE TRUE
    WIN32_EXECUTABLE TRUE
)

target_compile_definitions(slater
    PRIVATE $<$<OR:$<CONFIG:Debug>,$<CONFIG:RelWithDebInfo>>:QT_QML_DEBUG>)
target_link_libraries(slater
    PRIVATE Qt6::Core Qt6::Quick Qt6::QuickControls2)
target_link_libraries(slater PRIVATE Qt6::WebView)
