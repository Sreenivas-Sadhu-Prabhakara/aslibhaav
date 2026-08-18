import 'package:flutter/material.dart';

void main() => runApp(const AslibhaavApp());

/// Aslibhaav — "the real price". Enter a distributor's scheme and see the true
/// effective cost per unit. The calculation mirrors the Go middleware's
/// EffectiveCostPerUnit; the Go service owns Postgres-backed history.
class AslibhaavApp extends StatelessWidget {
  const AslibhaavApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Aslibhaav',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: const Color(0xFF16733E),
        useMaterial3: true,
      ),
      home: const HomePage(),
    );
  }
}

/// Result of an effective-cost computation, with every intermediate value.
class Breakdown {
  final double afterTrade, afterCash, afterFreeGoods, freightPerUnit, effective;
  const Breakdown(this.afterTrade, this.afterCash, this.afterFreeGoods,
      this.freightPerUnit, this.effective);
}

/// effectiveCostPerUnit mirrors backend/cost.go exactly.
Breakdown effectiveCost({
  required double basePrice,
  required double tradePct,
  required double cashPct,
  required int buyQty,
  required int getQty,
  required double freight,
  required bool freightExtra,
}) {
  final afterTrade = basePrice * (1 - tradePct / 100);
  final afterCash = afterTrade * (1 - cashPct / 100);
  final unitsReceived = buyQty + getQty;
  var factor = 1.0;
  if (buyQty > 0 && unitsReceived > 0) factor = buyQty / unitsReceived;
  final afterFreeGoods = afterCash * factor;
  var freightPerUnit = 0.0;
  if (freightExtra) {
    final denom = unitsReceived <= 0 ? 1 : unitsReceived;
    freightPerUnit = freight / denom;
  }
  final effective = afterFreeGoods + freightPerUnit;
  return Breakdown(
      afterTrade, afterCash, afterFreeGoods, freightPerUnit, effective);
}

class HomePage extends StatefulWidget {
  const HomePage({super.key});
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final _base = TextEditingController(text: '100');
  final _trade = TextEditingController(text: '10');
  final _cash = TextEditingController(text: '5');
  final _buy = TextEditingController(text: '10');
  final _get = TextEditingController(text: '2');
  final _freight = TextEditingController(text: '120');
  bool _freightExtra = true;
  Breakdown? _result;

  double _n(TextEditingController c) => double.tryParse(c.text.trim()) ?? 0;
  int _i(TextEditingController c) => int.tryParse(c.text.trim()) ?? 0;

  void _compute() {
    setState(() {
      _result = effectiveCost(
        basePrice: _n(_base),
        tradePct: _n(_trade),
        cashPct: _n(_cash),
        buyQty: _i(_buy),
        getQty: _i(_get),
        freight: _n(_freight),
        freightExtra: _freightExtra,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    String money(double v) => '₹${v.toStringAsFixed(2)}';
    return Scaffold(
      appBar: AppBar(
        title: const Text('Aslibhaav · the real price'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text('Enter the distributor\'s scheme',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          _field(_base, 'Base price per unit (₹)'),
          _field(_trade, 'Trade discount %'),
          _field(_cash, 'Cash discount %'),
          Row(children: [
            Expanded(child: _field(_buy, 'Buy qty (X)')),
            const SizedBox(width: 12),
            Expanded(child: _field(_get, 'Free qty (Y)')),
          ]),
          _field(_freight, 'Freight (₹, total)'),
          SwitchListTile(
            title: const Text('Freight charged extra (on top)'),
            value: _freightExtra,
            onChanged: (v) => setState(() => _freightExtra = v),
            contentPadding: EdgeInsets.zero,
          ),
          const SizedBox(height: 8),
          FilledButton.icon(
            onPressed: _compute,
            icon: const Icon(Icons.calculate),
            label: const Text('Show the real price'),
          ),
          const SizedBox(height: 20),
          if (_result != null) _resultCard(money),
        ],
      ),
    );
  }

  Widget _field(TextEditingController c, String label) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: TextField(
          controller: c,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: InputDecoration(labelText: label, border: const OutlineInputBorder()),
          onChanged: (_) => _compute(),
        ),
      );

  Widget _resultCard(String Function(double) money) {
    final r = _result!;
    return Card(
      color: Theme.of(context).colorScheme.primaryContainer,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('True effective cost per unit',
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 4),
          Text(money(r.effective),
              style: const TextStyle(fontSize: 34, fontWeight: FontWeight.bold)),
          const Divider(height: 24),
          _row('After trade discount', money(r.afterTrade)),
          _row('After cash discount', money(r.afterCash)),
          _row('After free goods', money(r.afterFreeGoods)),
          _row('Freight per unit', money(r.freightPerUnit)),
        ]),
      ),
    );
  }

  Widget _row(String k, String v) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 2),
        child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [Text(k), Text(v, style: const TextStyle(fontWeight: FontWeight.w600))]),
      );
}
