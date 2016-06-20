/****************************************************************************
** Meta object code from reading C++ file 'qml-regexp.h'
**
** Created by: The Qt Meta Object Compiler version 67 (Qt 5.5.0)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "qml-regexp.h"
#include <QtCore/qbytearray.h>
#include <QtCore/qmetatype.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'qml-regexp.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 67
#error "This file was generated using the moc from 5.5.0. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

QT_BEGIN_MOC_NAMESPACE
struct qt_meta_stringdata_RegExp_t {
    QByteArrayData data[11];
    char stringdata0[116];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_RegExp_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_RegExp_t qt_meta_stringdata_RegExp = {
    {
QT_MOC_LITERAL(0, 0, 6), // "RegExp"
QT_MOC_LITERAL(1, 7, 15), // "validityChanged"
QT_MOC_LITERAL(2, 23, 0), // ""
QT_MOC_LITERAL(3, 24, 14), // "patternChanged"
QT_MOC_LITERAL(4, 39, 12), // "errorChanged"
QT_MOC_LITERAL(5, 52, 22), // "caseSensitivityChanged"
QT_MOC_LITERAL(6, 75, 7), // "pattern"
QT_MOC_LITERAL(7, 83, 5), // "valid"
QT_MOC_LITERAL(8, 89, 5), // "error"
QT_MOC_LITERAL(9, 95, 6), // "regexp"
QT_MOC_LITERAL(10, 102, 13) // "caseSensitive"

    },
    "RegExp\0validityChanged\0\0patternChanged\0"
    "errorChanged\0caseSensitivityChanged\0"
    "pattern\0valid\0error\0regexp\0caseSensitive"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_RegExp[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       4,   14, // methods
       5,   46, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       4,       // signalCount

 // signals: name, argc, parameters, tag, flags
       1,    1,   34,    2, 0x06 /* Public */,
       3,    1,   37,    2, 0x06 /* Public */,
       4,    1,   40,    2, 0x06 /* Public */,
       5,    1,   43,    2, 0x06 /* Public */,

 // signals: parameters
    QMetaType::Void, QMetaType::Bool,    2,
    QMetaType::Void, QMetaType::QString,    2,
    QMetaType::Void, QMetaType::QString,    2,
    QMetaType::Void, QMetaType::Bool,    2,

 // properties: name, type, flags
       6, QMetaType::QString, 0x00495103,
       7, QMetaType::Bool, 0x00495001,
       8, QMetaType::QString, 0x00495001,
       9, QMetaType::QRegExp, 0x00495001,
      10, QMetaType::Bool, 0x00495103,

 // properties: notify_signal_id
       1,
       0,
       2,
       1,
       3,

       0        // eod
};

void RegExp::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        RegExp *_t = static_cast<RegExp *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: _t->validityChanged((*reinterpret_cast< bool(*)>(_a[1]))); break;
        case 1: _t->patternChanged((*reinterpret_cast< const QString(*)>(_a[1]))); break;
        case 2: _t->errorChanged((*reinterpret_cast< const QString(*)>(_a[1]))); break;
        case 3: _t->caseSensitivityChanged((*reinterpret_cast< bool(*)>(_a[1]))); break;
        default: ;
        }
    } else if (_c == QMetaObject::IndexOfMethod) {
        int *result = reinterpret_cast<int *>(_a[0]);
        void **func = reinterpret_cast<void **>(_a[1]);
        {
            typedef void (RegExp::*_t)(bool );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&RegExp::validityChanged)) {
                *result = 0;
            }
        }
        {
            typedef void (RegExp::*_t)(const QString & );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&RegExp::patternChanged)) {
                *result = 1;
            }
        }
        {
            typedef void (RegExp::*_t)(const QString & );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&RegExp::errorChanged)) {
                *result = 2;
            }
        }
        {
            typedef void (RegExp::*_t)(bool );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&RegExp::caseSensitivityChanged)) {
                *result = 3;
            }
        }
    }
#ifndef QT_NO_PROPERTIES
    else if (_c == QMetaObject::ReadProperty) {
        RegExp *_t = static_cast<RegExp *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: *reinterpret_cast< QString*>(_v) = _t->pattern(); break;
        case 1: *reinterpret_cast< bool*>(_v) = _t->isValid(); break;
        case 2: *reinterpret_cast< QString*>(_v) = _t->errorString(); break;
        case 3: *reinterpret_cast< QRegExp*>(_v) = _t->Internal(); break;
        case 4: *reinterpret_cast< bool*>(_v) = _t->caseSensitive(); break;
        default: break;
        }
    } else if (_c == QMetaObject::WriteProperty) {
        RegExp *_t = static_cast<RegExp *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: _t->setPattern(*reinterpret_cast< QString*>(_v)); break;
        case 4: _t->setCaseSensitive(*reinterpret_cast< bool*>(_v)); break;
        default: break;
        }
    } else if (_c == QMetaObject::ResetProperty) {
    }
#endif // QT_NO_PROPERTIES
}

const QMetaObject RegExp::staticMetaObject = {
    { &QObject::staticMetaObject, qt_meta_stringdata_RegExp.data,
      qt_meta_data_RegExp,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *RegExp::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *RegExp::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_RegExp.stringdata0))
        return static_cast<void*>(const_cast< RegExp*>(this));
    return QObject::qt_metacast(_clname);
}

int RegExp::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
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
        _id -= 5;
    } else if (_c == QMetaObject::QueryPropertyDesignable) {
        _id -= 5;
    } else if (_c == QMetaObject::QueryPropertyScriptable) {
        _id -= 5;
    } else if (_c == QMetaObject::QueryPropertyStored) {
        _id -= 5;
    } else if (_c == QMetaObject::QueryPropertyEditable) {
        _id -= 5;
    } else if (_c == QMetaObject::QueryPropertyUser) {
        _id -= 5;
    }
#endif // QT_NO_PROPERTIES
    return _id;
}

// SIGNAL 0
void RegExp::validityChanged(bool _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 0, _a);
}

// SIGNAL 1
void RegExp::patternChanged(const QString & _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 1, _a);
}

// SIGNAL 2
void RegExp::errorChanged(const QString & _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 2, _a);
}

// SIGNAL 3
void RegExp::caseSensitivityChanged(bool _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 3, _a);
}
QT_END_MOC_NAMESPACE
