/****************************************************************************
** Meta object code from reading C++ file 'progressbar.h'
**
** Created by: The Qt Meta Object Compiler version 67 (Qt 5.5.0)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "progressbar.h"
#include <QtCore/qbytearray.h>
#include <QtCore/qmetatype.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'progressbar.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 67
#error "This file was generated using the moc from 5.5.0. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

QT_BEGIN_MOC_NAMESPACE
struct qt_meta_stringdata_ProgressBar_t {
    QByteArrayData data[15];
    char stringdata0[205];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ProgressBar_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ProgressBar_t qt_meta_stringdata_ProgressBar = {
    {
QT_MOC_LITERAL(0, 0, 11), // "ProgressBar"
QT_MOC_LITERAL(1, 12, 20), // "indeterminateChanged"
QT_MOC_LITERAL(2, 33, 0), // ""
QT_MOC_LITERAL(3, 34, 19), // "maximumValueChanged"
QT_MOC_LITERAL(4, 54, 19), // "minimumValueChanged"
QT_MOC_LITERAL(5, 74, 18), // "orientationChanged"
QT_MOC_LITERAL(6, 93, 15), // "Qt::Orientation"
QT_MOC_LITERAL(7, 109, 12), // "valueChanged"
QT_MOC_LITERAL(8, 122, 15), // "selectedChanged"
QT_MOC_LITERAL(9, 138, 13), // "indeterminate"
QT_MOC_LITERAL(10, 152, 12), // "maximumValue"
QT_MOC_LITERAL(11, 165, 12), // "minimumValue"
QT_MOC_LITERAL(12, 178, 11), // "orientation"
QT_MOC_LITERAL(13, 190, 5), // "value"
QT_MOC_LITERAL(14, 196, 8) // "selected"

    },
    "ProgressBar\0indeterminateChanged\0\0"
    "maximumValueChanged\0minimumValueChanged\0"
    "orientationChanged\0Qt::Orientation\0"
    "valueChanged\0selectedChanged\0indeterminate\0"
    "maximumValue\0minimumValue\0orientation\0"
    "value\0selected"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ProgressBar[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       6,   14, // methods
       6,   62, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       6,       // signalCount

 // signals: name, argc, parameters, tag, flags
       1,    1,   44,    2, 0x06 /* Public */,
       3,    1,   47,    2, 0x06 /* Public */,
       4,    1,   50,    2, 0x06 /* Public */,
       5,    1,   53,    2, 0x06 /* Public */,
       7,    1,   56,    2, 0x06 /* Public */,
       8,    1,   59,    2, 0x06 /* Public */,

 // signals: parameters
    QMetaType::Void, QMetaType::Bool,    2,
    QMetaType::Void, QMetaType::Double,    2,
    QMetaType::Void, QMetaType::Double,    2,
    QMetaType::Void, 0x80000000 | 6,    2,
    QMetaType::Void, QMetaType::Double,    2,
    QMetaType::Void, QMetaType::Bool,    2,

 // properties: name, type, flags
       9, QMetaType::Bool, 0x00495103,
      10, QMetaType::Double, 0x00495103,
      11, QMetaType::Double, 0x00495103,
      12, 0x80000000 | 6, 0x0049510b,
      13, QMetaType::Double, 0x00495103,
      14, QMetaType::Bool, 0x00495103,

 // properties: notify_signal_id
       0,
       1,
       2,
       3,
       4,
       5,

       0        // eod
};

void ProgressBar::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        ProgressBar *_t = static_cast<ProgressBar *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: _t->indeterminateChanged((*reinterpret_cast< bool(*)>(_a[1]))); break;
        case 1: _t->maximumValueChanged((*reinterpret_cast< double(*)>(_a[1]))); break;
        case 2: _t->minimumValueChanged((*reinterpret_cast< double(*)>(_a[1]))); break;
        case 3: _t->orientationChanged((*reinterpret_cast< Qt::Orientation(*)>(_a[1]))); break;
        case 4: _t->valueChanged((*reinterpret_cast< double(*)>(_a[1]))); break;
        case 5: _t->selectedChanged((*reinterpret_cast< bool(*)>(_a[1]))); break;
        default: ;
        }
    } else if (_c == QMetaObject::IndexOfMethod) {
        int *result = reinterpret_cast<int *>(_a[0]);
        void **func = reinterpret_cast<void **>(_a[1]);
        {
            typedef void (ProgressBar::*_t)(bool );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::indeterminateChanged)) {
                *result = 0;
            }
        }
        {
            typedef void (ProgressBar::*_t)(double );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::maximumValueChanged)) {
                *result = 1;
            }
        }
        {
            typedef void (ProgressBar::*_t)(double );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::minimumValueChanged)) {
                *result = 2;
            }
        }
        {
            typedef void (ProgressBar::*_t)(Qt::Orientation );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::orientationChanged)) {
                *result = 3;
            }
        }
        {
            typedef void (ProgressBar::*_t)(double );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::valueChanged)) {
                *result = 4;
            }
        }
        {
            typedef void (ProgressBar::*_t)(bool );
            if (*reinterpret_cast<_t *>(func) == static_cast<_t>(&ProgressBar::selectedChanged)) {
                *result = 5;
            }
        }
    }
#ifndef QT_NO_PROPERTIES
    else if (_c == QMetaObject::ReadProperty) {
        ProgressBar *_t = static_cast<ProgressBar *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: *reinterpret_cast< bool*>(_v) = _t->indeterminate(); break;
        case 1: *reinterpret_cast< double*>(_v) = _t->maximumValue(); break;
        case 2: *reinterpret_cast< double*>(_v) = _t->minimumValue(); break;
        case 3: *reinterpret_cast< Qt::Orientation*>(_v) = _t->orientation(); break;
        case 4: *reinterpret_cast< double*>(_v) = _t->value(); break;
        case 5: *reinterpret_cast< bool*>(_v) = _t->selected(); break;
        default: break;
        }
    } else if (_c == QMetaObject::WriteProperty) {
        ProgressBar *_t = static_cast<ProgressBar *>(_o);
        void *_v = _a[0];
        switch (_id) {
        case 0: _t->setIndeterminate(*reinterpret_cast< bool*>(_v)); break;
        case 1: _t->setMaximumValue(*reinterpret_cast< double*>(_v)); break;
        case 2: _t->setMinimumValue(*reinterpret_cast< double*>(_v)); break;
        case 3: _t->setOrientation(*reinterpret_cast< Qt::Orientation*>(_v)); break;
        case 4: _t->setValue(*reinterpret_cast< double*>(_v)); break;
        case 5: _t->setSelected(*reinterpret_cast< bool*>(_v)); break;
        default: break;
        }
    } else if (_c == QMetaObject::ResetProperty) {
    }
#endif // QT_NO_PROPERTIES
}

const QMetaObject ProgressBar::staticMetaObject = {
    { &QQuickPaintedItem::staticMetaObject, qt_meta_stringdata_ProgressBar.data,
      qt_meta_data_ProgressBar,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ProgressBar::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ProgressBar::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ProgressBar.stringdata0))
        return static_cast<void*>(const_cast< ProgressBar*>(this));
    return QQuickPaintedItem::qt_metacast(_clname);
}

int ProgressBar::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QQuickPaintedItem::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 6)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 6;
    } else if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 6)
            *reinterpret_cast<int*>(_a[0]) = -1;
        _id -= 6;
    }
#ifndef QT_NO_PROPERTIES
   else if (_c == QMetaObject::ReadProperty || _c == QMetaObject::WriteProperty
            || _c == QMetaObject::ResetProperty || _c == QMetaObject::RegisterPropertyMetaType) {
        qt_static_metacall(this, _c, _id, _a);
        _id -= 6;
    } else if (_c == QMetaObject::QueryPropertyDesignable) {
        _id -= 6;
    } else if (_c == QMetaObject::QueryPropertyScriptable) {
        _id -= 6;
    } else if (_c == QMetaObject::QueryPropertyStored) {
        _id -= 6;
    } else if (_c == QMetaObject::QueryPropertyEditable) {
        _id -= 6;
    } else if (_c == QMetaObject::QueryPropertyUser) {
        _id -= 6;
    }
#endif // QT_NO_PROPERTIES
    return _id;
}

// SIGNAL 0
void ProgressBar::indeterminateChanged(bool _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 0, _a);
}

// SIGNAL 1
void ProgressBar::maximumValueChanged(double _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 1, _a);
}

// SIGNAL 2
void ProgressBar::minimumValueChanged(double _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 2, _a);
}

// SIGNAL 3
void ProgressBar::orientationChanged(Qt::Orientation _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 3, _a);
}

// SIGNAL 4
void ProgressBar::valueChanged(double _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 4, _a);
}

// SIGNAL 5
void ProgressBar::selectedChanged(bool _t1)
{
    void *_a[] = { Q_NULLPTR, const_cast<void*>(reinterpret_cast<const void*>(&_t1)) };
    QMetaObject::activate(this, &staticMetaObject, 5, _a);
}
QT_END_MOC_NAMESPACE
