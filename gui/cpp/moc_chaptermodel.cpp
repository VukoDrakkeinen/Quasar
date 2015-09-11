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
    QByteArrayData data[11];
    char stringdata0[96];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ChapterModel_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ChapterModel_t qt_meta_stringdata_ChapterModel = {
    {
QT_MOC_LITERAL(0, 0, 12), // "ChapterModel"
QT_MOC_LITERAL(1, 13, 15), // "comicIdxChanged"
QT_MOC_LITERAL(2, 29, 0), // ""
QT_MOC_LITERAL(3, 30, 7), // "comicId"
QT_MOC_LITERAL(4, 38, 9), // "ccomicIdx"
QT_MOC_LITERAL(5, 48, 11), // "setComicIdx"
QT_MOC_LITERAL(6, 60, 8), // "comicIdx"
QT_MOC_LITERAL(7, 69, 6), // "qmlGet"
QT_MOC_LITERAL(8, 76, 3), // "row"
QT_MOC_LITERAL(9, 80, 6), // "column"
QT_MOC_LITERAL(10, 87, 8) // "roleName"

    },
    "ChapterModel\0comicIdxChanged\0\0comicId\0"
    "ccomicIdx\0setComicIdx\0comicIdx\0qmlGet\0"
    "row\0column\0roleName"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ChapterModel[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       4,   14, // methods
       1,   48, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       1,       // signalCount

 // signals: name, argc, parameters, tag, flags
       1,    1,   34,    2, 0x06 /* Public */,

 // methods: name, argc, parameters, tag, flags
       4,    0,   37,    2, 0x02 /* Public */,
       5,    1,   38,    2, 0x02 /* Public */,
       7,    3,   41,    2, 0x02 /* Public */,

 // signals: parameters
    QMetaType::Void, QMetaType::Int,    3,

 // methods: parameters
    QMetaType::Int,
    QMetaType::Void, QMetaType::Int,    6,
    QMetaType::QVariant, QMetaType::Int, QMetaType::Int, QMetaType::QString,    8,    9,   10,

 // properties: name, type, flags
       3, QMetaType::Int, 0x00495003,

 // properties: notify_signal_id
       0,

       0        // eod
};

void ChapterModel::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        ChapterModel *_t = static_cast<ChapterModel *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: _t->comicIdxChanged((*reinterpret_cast< int(*)>(_a[1]))); break;
        case 1: { int _r = _t->ccomicIdx();
            if (_a[0]) *reinterpret_cast< int*>(_a[0]) = _r; }  break;
        case 2: _t->setComicIdx((*reinterpret_cast< int(*)>(_a[1]))); break;
        case 3: { QVariant _r = _t->qmlGet((*reinterpret_cast< int(*)>(_a[1])),(*reinterpret_cast< int(*)>(_a[2])),(*reinterpret_cast< const QString(*)>(_a[3])));
            if (_a[0]) *reinterpret_cast< QVariant*>(_a[0]) = _r; }  break;
        default: ;
        }
    } else if (_c == QMetaObject::IndexOfMethod) {
        int *result = reinterpret_cast<int *>(_a[0]);
        void **func = reinterpret_cast<void **>(_a[1]);
        {
            typedef void (ChapterModel::*_t)(int );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ChapterModel::comicIdxChanged)) {
                *result = 0;
            }
        }
    }
#ifndef QT_NO_PROPERTIES
    else if (_c == QMetaObject::ReadProperty) {
        ChapterModel *_t = static_cast<ChapterModel *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: *reinterpret_cast< int*>(_v) = _t->ccomicIdx(); break;
        default: break;
        }
    } else if (_c == QMetaObject::WriteProperty) {
        ChapterModel *_t = static_cast<ChapterModel *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: _t->setComicIdx(*reinterpret_cast< int*>(_v)); break;
        default: break;
        }
    } else if (_c == QMetaObject::ResetProperty) {
    }
#endif // QT_NO_PROPERTIES
}

const QMetaObject ChapterModel::staticMetaObject = {
    { &NotifiableModel::staticMetaObject, qt_meta_stringdata_ChapterModel.data,
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
    return NotifiableModel::qt_metacast(_clname);
}

int ChapterModel::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = NotifiableModel::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 4)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 4;
    } else if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 4)
            *reinterpret_cast<int*>(_a[0]) = -1;
        _id -= 4;
    }
#ifndef QT_NO_PROPERTIES
   else if (_c == QMetaObject::ReadProperty || _c == QMetaObject::WriteProperty
            || _c == QMetaObject::ResetProperty || _c == QMetaObject::RegisterPropertyMetaType) {
        qt_static_metacall(this, _c, _id, _a);
        _id -= 1;
    } else if (_c == QMetaObject::QueryPropertyDesignable) {
        _id -= 1;
    } else if (_c == QMetaObject::QueryPropertyScriptable) {
        _id -= 1;
    } else if (_c == QMetaObject::QueryPropertyStored) {
        _id -= 1;
    } else if (_c == QMetaObject::QueryPropertyEditable) {
        _id -= 1;
    } else if (_c == QMetaObject::QueryPropertyUser) {
        _id -= 1;
    }
#endif // QT_NO_PROPERTIES
    return _id;
}

// SIGNAL 0
void ChapterModel::comicIdxChanged(int _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 0, _a);
}
QT_END_MOC_NAMESPACE
