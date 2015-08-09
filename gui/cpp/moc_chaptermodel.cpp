/****************************************************************************
** Meta object code from reading C++ file 'chaptermodel.h'
**
** Created by: The Qt Meta Object Compiler version 67 (Qt 5.5.0)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "chaptermodel.h"
#include <QtCore/qbytearray.h>
#include <QtCore/qmetatype.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'chaptermodel.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 67
#error "This file was generated using the moc from 5.5.0. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

QT_BEGIN_MOC_NAMESPACE
struct qt_meta_stringdata_ChapterModel_t {
    QByteArrayData data[8];
    char stringdata0[62];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ChapterModel_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ChapterModel_t qt_meta_stringdata_ChapterModel = {
    {
QT_MOC_LITERAL(0, 0, 12), // "ChapterModel"
QT_MOC_LITERAL(1, 13, 11), // "setComicIdx"
QT_MOC_LITERAL(2, 25, 0), // ""
QT_MOC_LITERAL(3, 26, 8), // "comicIdx"
QT_MOC_LITERAL(4, 35, 6), // "qmlGet"
QT_MOC_LITERAL(5, 42, 3), // "row"
QT_MOC_LITERAL(6, 46, 6), // "column"
QT_MOC_LITERAL(7, 53, 8) // "roleName"

    },
    "ChapterModel\0setComicIdx\0\0comicIdx\0"
    "qmlGet\0row\0column\0roleName"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ChapterModel[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       2,   14, // methods
       0,    0, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // methods: name, argc, parameters, tag, flags
       1,    1,   24,    2, 0x02 /* Public */,
       4,    3,   27,    2, 0x02 /* Public */,

 // methods: parameters
    QMetaType::Void, QMetaType::Int,    3,
    QMetaType::QVariant, QMetaType::Int, QMetaType::Int, QMetaType::QString,    5,    6,    7,

       0        // eod
};

void ChapterModel::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        ChapterModel *_t = static_cast<ChapterModel *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: _t->setComicIdx((*reinterpret_cast< int(*)>(_a[1]))); break;
        case 1: { QVariant _r = _t->qmlGet((*reinterpret_cast< int(*)>(_a[1])),(*reinterpret_cast< int(*)>(_a[2])),(*reinterpret_cast< const QString(*)>(_a[3])));
            if (_a[0]) *reinterpret_cast< QVariant*>(_a[0]) = _r; }  break;
        default: ;
        }
    }
}

const QMetaObject ChapterModel::staticMetaObject = {
    { &QAbstractItemModel::staticMetaObject, qt_meta_stringdata_ChapterModel.data,
      qt_meta_data_ChapterModel,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ChapterModel::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ChapterModel::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ChapterModel.stringdata0))
        return static_cast<void*>(const_cast< ChapterModel*>(this));
    return QAbstractItemModel::qt_metacast(_clname);
}

int ChapterModel::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QAbstractItemModel::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 2)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 2;
    } else if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 2)
            *reinterpret_cast<int*>(_a[0]) = -1;
        _id -= 2;
    }
    return _id;
}
QT_END_MOC_NAMESPACE
