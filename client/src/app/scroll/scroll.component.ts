import { AfterViewInit, Component, ElementRef, OnDestroy, ViewChildren } from '@angular/core';
import { Subscription } from 'rxjs/Subscription';
import { Point } from './point';

// https://github.com/angular/angular/issues/8663
@Component({
  templateUrl: './scroll.component.html',
  styleUrls: ['./scroll.component.css'],
})
export class ScrollComponent implements AfterViewInit, OnDestroy {
  @ViewChildren('pointListEntry') pointListEntry: any;
  pointListEntryChangeSubscription: Subscription;
  pointList: Array<Point> = [];
  activePoint = 0;
  activePosition = 0.0;
  

  constructor(public elementRef: ElementRef) {
  }

  calculateActivePosition() {
    let scroller: HTMLElement = this.elementRef.nativeElement.querySelector('.scroller');
    let active = scroller.querySelector('.active-point');
    if (!active) {
      this.activePosition = 0;
    } else {
      this.activePosition = active.getBoundingClientRect().top - scroller.getBoundingClientRect().top + scroller.scrollTop;
    }
  }

  ngAfterViewInit() {
    // For the first time
    this.onPointListEntryChanged();

    // For next times
    this.pointListEntryChangeSubscription = this.pointListEntry.changes.subscribe(
      () => {this.onPointListEntryChanged()}
    );
  }

  ngOnDestroy() {
    if (this.pointListEntryChangeSubscription) {
      this.pointListEntryChangeSubscription.unsubscribe();
      this.pointListEntryChangeSubscription = null;
    }
  }

  onPointListEntryChanged() {
    // ���݃A�N�e�B�u�ȃZ���ɃX�N���[������
    // Angular ���̂ɂ̓X�N���[���𐧌䂷��@�\���Ȃ����߁A
    // DOM �ɒ��ڃA�N�Z�X���Ď�������B
    let scroller: HTMLElement = this.elementRef.nativeElement.querySelector('.scroller');
    let active = scroller.querySelector('.active-point');
    if (!active) {
      scroller.scrollTop = 0;
      return;
    }

    let activeRect = active.getBoundingClientRect();
    let scrollerRect = scroller.getBoundingClientRect();
    // scroller ���ł̃A�N�e�B�u�Z���̑��Έʒu
    // �A�N�e�B�u�Z���̃y�[�W���ʒu - scroller �̃y�[�W���ʒu + ���݂̃X�N���[���ʒu
    let activeTop = activeRect.top - scrollerRect.top + scroller.scrollTop;

    // activeTop �ɃX�N���[������ƃA�N�e�B�u�Z������ԏ�ɗ����ԂɃX�N���[������B
    // �^�񒆂ɕ\���������̂ŁA�ȉ��̕␳���s��
    // + scroller �̍��� / 2
    // - �Z���̍��� / 2
    let scrollTarget = activeTop - (scroller.clientHeight / 2) + ((activeRect.bottom - activeRect.top) / 2);

    scroller.scrollTop = scrollTarget;
  }

  generatePoints() {
    let newPointList: Array<Point> = [];
    let size = 30;
    let active = Math.floor(Math.random() * size) + 1;
    let now = new Date();
    for (let i = 0; i < size; ++i) {
      let point = new Point();
      point.point = i + 1;
      if (i <= active) {
        point.createdAt = new Date(now.getTime() + (i * 24 * 60 * 60));
      }
      newPointList.push(point);
    }
    this.pointList = newPointList;
    this.activePoint = active;
  }
}
